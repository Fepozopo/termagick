package internal

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// previewHistogramFromWand computes per-channel equalized histograms from the provided wand,
// renders them to a PNG via createHistogramPNG and previews (or writes a temp PNG on failure).
func previewHistogramFromWand(wand *imagick.MagickWand, bins int) error {
	if wand == nil {
		return fmt.Errorf("nil wand")
	}

	// Export full image pixels as RGBA (PIXEL_CHAR yields 0-255 values).
	w := int(wand.GetImageWidth())
	h := int(wand.GetImageHeight())
	if w == 0 || h == 0 {
		return fmt.Errorf("image has zero dimensions")
	}
	pixIface, err := wand.ExportImagePixels(0, 0, uint(w), uint(h), "RGBA", imagick.PIXEL_CHAR)
	if err != nil {
		return fmt.Errorf("ExportImagePixels failed: %w", err)
	}

	// Normalize pixel data to []byte
	var pixels []byte
	switch v := pixIface.(type) {
	case []byte:
		pixels = v
	case []uint16:
		pixels = make([]byte, len(v))
		for i := range v {
			pixels[i] = byte(v[i] >> 8)
		}
	case []float32:
		pixels = make([]byte, len(v))
		for i := range v {
			fv := v[i]
			if fv <= 0 {
				pixels[i] = 0
			} else if fv >= 1 {
				pixels[i] = 255
			} else {
				pixels[i] = byte(fv * 255.0)
			}
		}
	case []float64:
		pixels = make([]byte, len(v))
		for i := range v {
			fv := v[i]
			if fv <= 0 {
				pixels[i] = 0
			} else if fv >= 1 {
				pixels[i] = 255
			} else {
				pixels[i] = byte(fv * 255.0)
			}
		}
	default:
		return fmt.Errorf("unsupported pixel data type: %T", v)
	}
	if len(pixels) < 4 {
		return fmt.Errorf("no pixel data")
	}

	numPixels := len(pixels) / 4

	// Extract channel arrays.
	rVals := make([]uint8, numPixels)
	gVals := make([]uint8, numPixels)
	bVals := make([]uint8, numPixels)
	for i := 0; i < numPixels; i++ {
		o := i * 4
		rVals[i] = uint8(pixels[o])
		gVals[i] = uint8(pixels[o+1])
		bVals[i] = uint8(pixels[o+2])
	}

	// Function to compute histogram[256] for a channel
	hist256 := func(vals []uint8) []int {
		h := make([]int, 256)
		for _, v := range vals {
			h[int(v)]++
		}
		return h
	}

	// Compute per-channel histograms (256 bins)
	hR := hist256(rVals)
	hG := hist256(gVals)
	hB := hist256(bVals)

	// Compute equalization map for a 256-level channel histogram
	equalizeMap := func(h []int) [256]uint8 {
		total := 0
		for _, c := range h {
			total += c
		}
		var cmap [256]uint8
		if total == 0 {
			for i := 0; i < 256; i++ {
				cmap[i] = uint8(i)
			}
			return cmap
		}
		// CDF
		cdf := make([]int, 256)
		cdf[0] = h[0]
		for i := 1; i < 256; i++ {
			cdf[i] = cdf[i-1] + h[i]
		}
		// Find cdf_min (first non-zero)
		cdfMin := 0
		for i := 0; i < 256; i++ {
			if cdf[i] != 0 {
				cdfMin = cdf[i]
				break
			}
		}
		den := float64(total - cdfMin)
		if den <= 0 {
			// degenerate: map to identity
			for i := 0; i < 256; i++ {
				cmap[i] = uint8(i)
			}
			return cmap
		}
		for i := 0; i < 256; i++ {
			val := float64(cdf[i]-cdfMin) / den
			if val < 0 {
				val = 0
			} else if val > 1 {
				val = 1
			}
			cmap[i] = uint8(math.Round(val * 255.0))
		}
		return cmap
	}

	mapR := equalizeMap(hR)
	mapG := equalizeMap(hG)
	mapB := equalizeMap(hB)

	// Apply equalization maps to pixels to obtain equalized channel values.
	eqR := make([]uint8, numPixels)
	eqG := make([]uint8, numPixels)
	eqB := make([]uint8, numPixels)
	for i := 0; i < numPixels; i++ {
		eqR[i] = mapR[rVals[i]]
		eqG[i] = mapG[gVals[i]]
		eqB[i] = mapB[bVals[i]]
	}

	// Build histograms for equalized channels using requested bin count.
	histBins := func(vals []uint8, bins int) []int {
		out := make([]int, bins)
		for _, v := range vals {
			idx := int(v) * bins / 256
			if idx >= bins {
				idx = bins - 1
			}
			out[idx]++
		}
		return out
	}
	hREq := histBins(eqR, bins)
	hGEq := histBins(eqG, bins)
	hBEq := histBins(eqB, bins)

	// Render PNG via helper
	pngBytes, err := createHistogramPNG(bins, hREq, hGEq, hBEq)
	if err != nil {
		return err
	}

	// Preview via existing helper
	outWand := imagick.NewMagickWand()
	if outWand == nil {
		return fmt.Errorf("failed to create magick wand for histogram")
	}
	defer outWand.Destroy()
	if err := outWand.ReadImageBlob(pngBytes); err != nil {
		// As a fallback, write PNG to temp file so user can inspect it.
		tmp := os.TempDir() + "/termagick_histogram.png"
		if writeErr := os.WriteFile(tmp, pngBytes, 0644); writeErr == nil {
			return fmt.Errorf("failed to create magick image: %v (wrote PNG to %s)", err, tmp)
		} else {
			return fmt.Errorf("failed to create magick image: %v (also failed to write temp PNG: %v)", err, writeErr)
		}
	}

	// Try preview. If preview fails, write temp PNG and inform user.
	if err := PreviewWand(outWand); err != nil {
		tmp := os.TempDir() + "/termagick_histogram.png"
		writeErr := os.WriteFile(tmp, pngBytes, 0644)
		if writeErr == nil {
			fmt.Fprintf(os.Stderr, "Histogram written to %s (preview not supported or failed: %v)\n", tmp, err)
			return nil
		}
		return fmt.Errorf("preview failed: %v (also failed to write PNG: %v)", err, writeErr)
	}
	return nil
}

// createHistogramPNG renders histogram curves (R, G, B) into a PNG and returns the bytes.
// It accepts the number of bins and per-channel counts.
func createHistogramPNG(bins int, hREq, hGEq, hBEq []int) ([]byte, error) {
	// Prepare PNG canvas
	imgW := int(math.Max(640, float64(bins*3))) // ensure reasonably visible width
	imgH := 240
	left := 30
	right := 20
	top := 10
	bottom := 30
	plotW := imgW - left - right
	plotH := imgH - top - bottom

	canvas := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	// white background
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	// find global max for scaling
	maxCount := 1
	all := [][]int{hREq, hGEq, hBEq}
	for _, arr := range all {
		for _, v := range arr {
			if v > maxCount {
				maxCount = v
			}
		}
	}

	// draw line using simple Bresenham algorithm
	drawLine := func(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA) {
		dx := int(math.Abs(float64(x1 - x0)))
		dy := int(math.Abs(float64(y1 - y0)))
		sx := -1
		if x0 < x1 {
			sx = 1
		}
		sy := -1
		if y0 < y1 {
			sy = 1
		}
		errVal := dx - dy
		for {
			if x0 >= 0 && x0 < img.Bounds().Dx() && y0 >= 0 && y0 < img.Bounds().Dy() {
				img.SetRGBA(x0, y0, col)
			}
			if x0 == x1 && y0 == y1 {
				break
			}
			e2 := 2 * errVal
			if e2 > -dy {
				errVal -= dy
				x0 += sx
			}
			if e2 < dx {
				errVal += dx
				y0 += sy
			}
		}
	}

	// Function to plot a histogram curve given counts and color
	plotCurve := func(counts []int, col color.RGBA) {
		prevX, prevY := -1, -1
		for i := 0; i < bins; i++ {
			x := left
			if bins == 1 {
				x = left + plotW/2
			} else {
				x = left + int(math.Round(float64(i)*(float64(plotW-1)/float64(bins-1))))
			}
			val := counts[i]
			// scale val to plot height
			y := top + plotH - int(math.Round(float64(val)/float64(maxCount)*float64(plotH)))
			if y < top {
				y = top
			}
			if prevX >= 0 {
				drawLine(canvas, prevX, prevY, x, y, col)
			}
			prevX = x
			prevY = y
		}
	}

	// Plot R, G, B curves (alpha=255)
	plotCurve(hREq, color.RGBA{255, 64, 64, 255})
	plotCurve(hGEq, color.RGBA{64, 255, 64, 255})
	plotCurve(hBEq, color.RGBA{64, 64, 255, 255})

	// draw simple axes and labels
	axisColor := color.RGBA{0, 0, 0, 255}
	// x-axis
	drawLine(canvas, left, top+plotH, left+plotW-1, top+plotH, axisColor)
	// y-axis
	drawLine(canvas, left, top, left, top+plotH, axisColor)

	// legend boxes
	legendY := imgH - bottom + 6
	boxSize := 10
	// R
	draw.Draw(canvas, image.Rect(left, legendY, left+boxSize, legendY+boxSize), &image.Uniform{C: color.RGBA{255, 64, 64, 255}}, image.Point{}, draw.Src)
	// G
	draw.Draw(canvas, image.Rect(left+80, legendY, left+80+boxSize, legendY+boxSize), &image.Uniform{C: color.RGBA{64, 255, 64, 255}}, image.Point{}, draw.Src)
	// B
	draw.Draw(canvas, image.Rect(left+160, legendY, left+160+boxSize, legendY+boxSize), &image.Uniform{C: color.RGBA{64, 64, 255, 255}}, image.Point{}, draw.Src)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, fmt.Errorf("png encode failed: %w", err)
	}
	return buf.Bytes(), nil
}
