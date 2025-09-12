package internal

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

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
