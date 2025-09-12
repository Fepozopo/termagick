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
	"strconv"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// GetImageInfo returns a string with basic info about the image in the wand
func GetImageInfo(wand *imagick.MagickWand) (string, error) {
	if wand == nil {
		return "", fmt.Errorf("nil wand")
	}
	format := wand.GetImageFormat()
	width := wand.GetImageWidth()
	height := wand.GetImageHeight()
	compression := wand.GetImageCompression()
	compressionQuality := wand.GetImageCompressionQuality()

	// Resolve compression name using the shared mapping helper defined in meta.go.
	var compressionName string
	if name, ok := mapNumericToEnumName("compression", int64(compression)); ok {
		compressionName = name
	} else {
		// fallback to numeric representation if unknown
		compressionName = strconv.FormatInt(int64(compression), 10)
	}

	return fmt.Sprintf("Format: %s, Width: %d, Height: %d\nCompression: %s, Compression Quality: %v", format, width, height, compressionName, compressionQuality), nil
}

// ApplyCommand applies the given command to the magick wand
func ApplyCommand(wand *imagick.MagickWand, commandName string, args []string) error {
	switch commandName {
	case "adaptiveBlur":
		if len(args) != 2 {
			return fmt.Errorf("adaptiveBlur requires 2 arguments: radius and sigma")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		return wand.AdaptiveBlurImage(radius, sigma)

	case "adaptiveResize":
		if len(args) != 2 {
			return fmt.Errorf("adaptiveResize requires 2 arguments: columns and rows")
		}
		columns, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid columns: %w", err)
		}
		rows, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid rows: %w", err)
		}
		return wand.AdaptiveResizeImage(uint(columns), uint(rows))

	case "adaptiveSharpen":
		if len(args) != 2 {
			return fmt.Errorf("adaptiveSharpen requires 2 arguments: radius and sigma")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		return wand.AdaptiveSharpenImage(radius, sigma)

	case "adaptiveThreshold":
		if len(args) != 3 {
			return fmt.Errorf("adaptiveThreshold requires 3 arguments: width, height, and offset")
		}
		width, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid width: %w", err)
		}
		height, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid height: %w", err)
		}
		offset, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return fmt.Errorf("invalid offset: %w", err)
		}
		return wand.AdaptiveThresholdImage(uint(width), uint(height), offset)

	case "addNoise":
		if len(args) != 1 {
			return fmt.Errorf("addNoise requires 1 argument: noiseType")
		}
		noiseType, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid noiseType: %w", err)
		}
		return wand.AddNoiseImage(imagick.NoiseType(noiseType), 1)

	case "annotate":
		// annotate supports two forms:
		// 5 args: text, size, x, y, color
		// 6 args: text, font, size, x, y, color
		if !(len(args) == 5 || len(args) == 6) {
			return fmt.Errorf("annotate requires 5 or 6 arguments: text, [font], size, x, y, color")
		}
		text := args[0]
		font := ""
		sizeIdx := 1
		if len(args) == 6 {
			font = args[1]
			sizeIdx = 2
		}
		size, err := strconv.ParseFloat(args[sizeIdx], 64)
		if err != nil {
			return fmt.Errorf("invalid size: %w", err)
		}
		xFloat, err := strconv.ParseFloat(args[sizeIdx+1], 64)
		if err != nil {
			return fmt.Errorf("invalid x: %w", err)
		}
		yFloat, err := strconv.ParseFloat(args[sizeIdx+2], 64)
		if err != nil {
			return fmt.Errorf("invalid y: %w", err)
		}
		color := args[sizeIdx+3]

		dw := imagick.NewDrawingWand()
		defer dw.Destroy()
		if font != "" {
			dw.SetFont(font)
		}
		dw.SetFontSize(size)
		fill := imagick.NewPixelWand()
		defer fill.Destroy()
		fill.SetColor(color)
		dw.SetFillColor(fill)

		return wand.AnnotateImage(dw, xFloat, yFloat, 0.0, text)

	case "autoGamma":
		return wand.AutoGammaImage()

	case "autoLevel":
		return wand.AutoLevelImage()

	case "autoOrient":
		return wand.AutoOrientImage()

	case "blackThreshold":
		if len(args) != 1 {
			return fmt.Errorf("blackThreshold requires 1 argument: threshold")
		}
		pixel := imagick.NewPixelWand()
		defer pixel.Destroy()
		pixel.SetColor(args[0])
		return wand.BlackThresholdImage(pixel)

	case "blueShift":
		if len(args) != 1 {
			return fmt.Errorf("blueShift requires 1 argument: factor")
		}
		factor, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid factor: %w", err)
		}
		return wand.BlueShiftImage(factor)

	case "blur":
		if len(args) != 2 {
			return fmt.Errorf("blur requires 2 arguments: radius and sigma")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		return wand.BlurImage(radius, sigma)

	case "charcoal":
		if len(args) != 2 {
			return fmt.Errorf("charcoal requires 2 arguments: radius and sigma")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		return wand.CharcoalImage(radius, sigma)

	case "colorize":
		// colorize requires 2 args: color and opacity (0.0 - 1.0)
		if len(args) != 2 {
			return fmt.Errorf("colorize requires 2 arguments: color and opacity")
		}
		color := args[0]
		opacity, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid opacity: %w", err)
		}
		// Create pixel wand for colorize color
		colorPixel := imagick.NewPixelWand()
		defer colorPixel.Destroy()
		colorPixel.SetColor(color)

		// For opacity, use a pixel wand with an rgba alpha value.
		// Construct an rgba string with desired alpha; color channels are ignored for the opacity pixel.
		opacityPixel := imagick.NewPixelWand()
		defer opacityPixel.Destroy()
		// Clamp opacity to [0,1]
		if opacity < 0 {
			opacity = 0
		} else if opacity > 1 {
			opacity = 1
		}
		opacityPixel.SetColor(fmt.Sprintf("rgba(0,0,0,%f)", opacity))

		return wand.ColorizeImage(colorPixel, opacityPixel)

	case "composite":
		if len(args) != 4 {
			return fmt.Errorf("composite requires 4 arguments: sourceImagePath, composeOperator, x, y")
		}
		sourceWand := imagick.NewMagickWand()
		defer sourceWand.Destroy()
		// Read source image into its own wand
		if err := sourceWand.ReadImage(args[0]); err != nil {
			return fmt.Errorf("failed to read source image: %w", err)
		}
		compose, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid composeOperator: %w", err)
		}
		x, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid x: %w", err)
		}
		y, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid y: %w", err)
		}
		return wand.CompositeImage(sourceWand, imagick.CompositeOperator(compose), true, int(x), int(y))

	case "compress":
		// compress requires 2 args: type, quality
		if len(args) != 2 {
			return fmt.Errorf("compress requires 2 arguments: type and quality")
		}

		// Parse compression type
		compVal, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid compression type: %w", err)
		}

		// Set compression type
		if err := wand.SetImageCompression(imagick.CompressionType(compVal)); err != nil {
			return fmt.Errorf("failed to set image compression: %w", err)
		}

		// Parse quality
		q, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid quality: %w", err)
		}
		if q < 0 {
			q = 0
		}
		// Set compression quality
		if err := wand.SetImageCompressionQuality(uint(q)); err != nil {
			return fmt.Errorf("failed to set compression quality: %w", err)
		}
		return nil

	case "contrast":
		if len(args) != 1 {
			return fmt.Errorf("contrast requires 1 argument: sharpen (true/false)")
		}
		sharpen, err := strconv.ParseBool(args[0])
		if err != nil {
			return fmt.Errorf("invalid sharpen value: %w", err)
		}
		return wand.ContrastImage(sharpen)

	case "contrastStretch":
		// contrastStretch requires 2 args: low and high (floats)
		if len(args) != 2 {
			return fmt.Errorf("contrastStretch requires 2 arguments: low and high")
		}
		low, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid low value: %w", err)
		}
		high, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid high value: %w", err)
		}
		return wand.ContrastStretchImage(low, high)

	case "crop":
		// crop requires width, height, x, y
		if len(args) != 4 {
			return fmt.Errorf("crop requires 4 arguments: width, height, x, y")
		}
		width, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid width: %w", err)
		}
		height, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid height: %w", err)
		}
		x, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid x: %w", err)
		}
		y, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid y: %w", err)
		}
		return wand.CropImage(uint(width), uint(height), int(x), int(y))

	case "deskew":
		// deskew requires 1 arg: threshold
		if len(args) != 1 {
			return fmt.Errorf("deskew requires 1 argument: threshold")
		}
		threshold, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}
		return wand.DeskewImage(threshold)

	case "despeckle":
		// despeckle takes no args
		if len(args) != 0 {
			return fmt.Errorf("despeckle takes no arguments")
		}
		return wand.DespeckleImage()

	case "edge":
		if len(args) != 1 {
			return fmt.Errorf("edge requires 1 argument: radius")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		return wand.EdgeImage(radius)

	case "emboss":
		if len(args) != 2 {
			return fmt.Errorf("emboss requires 2 arguments: radius and sigma")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		return wand.EmbossImage(radius, sigma)

	case "equalize":
		return wand.EqualizeImage()

	case "enhance":
		return wand.EnhanceImage()

	case "flip":
		return wand.FlipImage()

	case "floodfillPaint":
		// floodfillPaint requires 6 args: fillColor, fuzz, borderColor, x, y, invert
		if len(args) != 6 {
			return fmt.Errorf("floodfillPaint requires 6 arguments: fillColor, fuzz, borderColor, x, y, invert")
		}
		fillColor := args[0]
		fuzz, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid fuzz: %w", err)
		}
		borderColor := args[2]
		x, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid x: %w", err)
		}
		y, err := strconv.ParseInt(args[4], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid y: %w", err)
		}
		invert, err := strconv.ParseBool(args[5])
		if err != nil {
			return fmt.Errorf("invalid invert value: %w", err)
		}
		// Prepare pixel wands
		fillPixel := imagick.NewPixelWand()
		defer fillPixel.Destroy()
		fillPixel.SetColor(fillColor)

		borderPixel := imagick.NewPixelWand()
		defer borderPixel.Destroy()
		borderPixel.SetColor(borderColor)

		return wand.FloodfillPaintImage(fillPixel, fuzz, borderPixel, int(x), int(y), invert)

	case "flop":
		return wand.FlopImage()

	case "gamma":
		if len(args) != 1 {
			return fmt.Errorf("gamma requires 1 argument: gamma")
		}
		gamma, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid gamma value: %w", err)
		}
		return wand.GammaImage(gamma)

	case "grayscale":
		return wand.SetImageColorspace(imagick.COLORSPACE_GRAY)

	case "histogram":
		// Equalize each RGB channel separately, then compute per-channel histograms
		// and render an overlaid-curve visualization (R in red, G in green, B in blue).
		// Optionally takes one argument: number of bins (default 256, max 4096).
		bins := 256
		if len(args) > 0 && args[0] != "" {
			if v, err := strconv.ParseInt(args[0], 10, 64); err == nil && v > 0 {
				if v > 4096 {
					bins = 4096
				} else {
					bins = int(v)
				}
			}
		}

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
			return fmt.Errorf("png encode failed: %w", err)
		}

		// Preview via existing helper
		outWand := imagick.NewMagickWand()
		if outWand == nil {
			return fmt.Errorf("failed to create magick wand for histogram")
		}
		defer outWand.Destroy()
		if err := outWand.ReadImageBlob(buf.Bytes()); err != nil {
			// As a fallback, write PNG to temp file so user can inspect it.
			tmp := os.TempDir() + "/termagick_histogram.png"
			if writeErr := os.WriteFile(tmp, buf.Bytes(), 0644); writeErr == nil {
				return fmt.Errorf("failed to create magick image: %v (wrote PNG to %s)", err, tmp)
			} else {
				return fmt.Errorf("failed to create magick image: %v (also failed to write temp PNG: %v)", err, writeErr)
			}
		}

		// Try preview. If preview fails, write temp PNG and inform user.
		if err := PreviewWand(outWand); err != nil {
			tmp := os.TempDir() + "/termagick_histogram.png"
			writeErr := os.WriteFile(tmp, buf.Bytes(), 0644)
			if writeErr == nil {
				fmt.Fprintf(os.Stderr, "Histogram written to %s (preview not supported or failed: %v)\n", tmp, err)
				return nil
			}
			return fmt.Errorf("preview failed: %v (also failed to write PNG: %v)", err, writeErr)
		}
		return nil

	case "identify":
		info := wand.IdentifyImage()
		fmt.Println(info)
		return nil

	case "level":
		if len(args) != 3 {
			return fmt.Errorf("level requires 3 arguments: blackPoint, gamma, whitePoint")
		}
		blackPoint, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid blackPoint: %w", err)
		}
		gamma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid gamma: %w", err)
		}
		whitePoint, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return fmt.Errorf("invalid whitePoint: %w", err)
		}
		return wand.LevelImage(blackPoint, gamma, whitePoint)

	case "medianFilter":
		if len(args) != 1 {
			return fmt.Errorf("medianFilter requires 1 argument: radius")
		}
		radius, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		return wand.StatisticImage(imagick.STATISTIC_MEDIAN, uint(radius), uint(radius))

	case "modulate":
		// modulate requires 3 args: brightness, saturation, hue
		if len(args) != 3 {
			return fmt.Errorf("modulate requires 3 arguments: brightness, saturation, hue")
		}
		brightness, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid brightness: %w", err)
		}
		saturation, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid saturation: %w", err)
		}
		hue, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return fmt.Errorf("invalid hue: %w", err)
		}
		return wand.ModulateImage(brightness, saturation, hue)

	case "monochrome":
		return wand.SetImageType(imagick.IMAGE_TYPE_BILEVEL)

	case "negate":
		if len(args) != 1 {
			return fmt.Errorf("negate requires 1 argument: only_gray (true/false)")
		}
		onlyGray, err := strconv.ParseBool(args[0])
		if err != nil {
			return fmt.Errorf("invalid only_gray value: %w", err)
		}
		return wand.NegateImage(onlyGray)

	case "normalize":
		return wand.NormalizeImage()

	case "oilpaint":
		if len(args) != 2 {
			return fmt.Errorf("oilpaint requires 2 arguments: radius and sigma")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		return wand.OilPaintImage(radius, sigma)

	case "posterize":
		if len(args) != 2 {
			return fmt.Errorf("posterize requires 2 arguments: levels and dither (true/false)")
		}
		levels, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid levels value: %w", err)
		}
		dither, err := strconv.ParseBool(args[1])
		if err != nil {
			return fmt.Errorf("invalid dither value: %w", err)
		}
		var ditherMethod imagick.DitherMethod
		if dither {
			ditherMethod = imagick.DITHER_METHOD_RIEMERSMA
		} else {
			ditherMethod = imagick.DITHER_METHOD_NO
		}
		return wand.PosterizeImage(uint(levels), ditherMethod)

	case "resize":
		if len(args) != 2 {
			return fmt.Errorf("resize requires 2 arguments: width and height")
		}
		width, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid width: %w", err)
		}
		height, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid height: %w", err)
		}
		return wand.ResizeImage(uint(width), uint(height), imagick.FILTER_LANCZOS)

	case "rotate":
		if len(args) != 1 {
			return fmt.Errorf("rotate requires 1 argument: degrees")
		}
		degrees, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid degrees: %w", err)
		}
		pixel := imagick.NewPixelWand()
		defer pixel.Destroy()
		pixel.SetColor("black")
		return wand.RotateImage(pixel, degrees)

	case "sepia":
		if len(args) != 1 {
			return fmt.Errorf("sepia requires 1 argument: percentage (0-100)")
		}
		percentage, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid percentage: %w", err)
		}
		if percentage < 0 || percentage > 100 {
			return fmt.Errorf("percentage must be between 0 and 100")
		}
		_, quantumRange := imagick.GetQuantumRange()
		threshold := percentage / 100 * float64(quantumRange)
		return wand.SepiaToneImage(threshold)

	case "sharpen":
		if len(args) != 2 {
			return fmt.Errorf("sharpen requires 2 arguments: radius and sigma")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		return wand.SharpenImage(radius, sigma)

	case "solarize":
		if len(args) != 1 {
			return fmt.Errorf("solarize requires 1 argument: threshold")
		}
		threshold, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}
		return wand.SolarizeImage(threshold)

	case "strip":
		// Remove image profiles and comments/metadata
		return wand.StripImage()

	case "swirl":
		if len(args) != 1 {
			return fmt.Errorf("swirl requires 1 argument: degrees")
		}
		degrees, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid degrees: %w", err)
		}
		return wand.SwirlImage(degrees, imagick.INTERPOLATE_PIXEL_BILINEAR)

	case "threshold":
		if len(args) != 1 {
			return fmt.Errorf("threshold requires 1 argument: threshold")
		}
		th, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid threshold value: %w", err)
		}
		return wand.ThresholdImage(th)

	case "trim":
		if len(args) != 1 {
			return fmt.Errorf("trim requires 1 argument: fuzz")
		}
		fuzz, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid fuzz value: %w", err)
		}
		return wand.TrimImage(fuzz)

	case "unsharp":
		if len(args) != 4 {
			return fmt.Errorf("unsharp requires 4 arguments: radius, sigma, amount, threshold")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		amount, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return fmt.Errorf("invalid amount: %w", err)
		}
		threshold, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}
		return wand.UnsharpMaskImage(radius, sigma, amount, threshold)

	case "vignette":
		if len(args) != 4 {
			return fmt.Errorf("vignette requires 4 arguments: radius, sigma, x, y")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		sigma, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid sigma: %w", err)
		}
		x, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid x: %w", err)
		}
		y, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid y: %w", err)
		}
		return wand.VignetteImage(radius, sigma, int(x), int(y))

	default:
		return fmt.Errorf("unknown command: %s", commandName)
	}
}
