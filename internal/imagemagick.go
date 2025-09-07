package internal

import (
	"fmt"
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
