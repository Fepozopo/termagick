package main

import (
	"fmt"
	"strconv"

	"gopkg.in/gographics/imagick.v3/imagick"
)

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

	case "composite":
		if len(args) != 4 {
			return fmt.Errorf("composite requires 4 arguments: sourceImagePath, composeOperator, x, y")
		}
		sourceWand := imagick.NewMagickWand()
		defer sourceWand.Destroy()
		err := sourceWand.ReadImage(args[0])
		if err != nil {
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

	case "contrast":
		if len(args) != 1 {
			return fmt.Errorf("contrast requires 1 argument: sharpen (true/false)")
		}
		sharpen, err := strconv.ParseBool(args[0])
		if err != nil {
			return fmt.Errorf("invalid sharpen value: %w", err)
		}
		return wand.ContrastImage(sharpen)

	case "edge":
		if len(args) != 1 {
			return fmt.Errorf("edge requires 1 argument: radius")
		}
		radius, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid radius: %w", err)
		}
		return wand.EdgeImage(radius)

	case "equalize":
		return wand.EqualizeImage()

	case "flip":
		return wand.FlipImage()

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
		// NewPixelWand creates a new pixel wand.
		pixel := imagick.NewPixelWand()
		// SetColor sets the color of the pixel wand.
		pixel.SetColor("black")
		return wand.RotateImage(pixel, degrees)

	case "sepia":
		if len(args) != 1 {
			return fmt.Errorf("sepia requires 1 argument: threshold")
		}
		threshold, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid threshold: %w", err)
		}
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

	case "trim":
		if len(args) != 1 {
			return fmt.Errorf("trim requires 1 argument: fuzz")
		}
		fuzz, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return fmt.Errorf("invalid fuzz value: %w", err)
		}
		return wand.TrimImage(fuzz)

	default:
		return fmt.Errorf("unknown command: %s", commandName)
	}
}
