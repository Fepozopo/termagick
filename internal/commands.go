package internal

// This file embeds the full command metadata directly in Go structs so the
// CLI can use a single compile-time source of truth.

var Commands = []CommandMeta{
	{
		Name:        "adaptiveBlur",
		Description: "Adaptively blur the image",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Neighborhood radius in pixels. Lower preserves finer detail; higher smooths larger areas.", Example: "1.0"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Standard deviation of the blur. Lower = subtle; higher = stronger smoothing.", Example: "2.0"},
		},
	},
	{
		Name:        "adaptiveResize",
		Description: "Adaptively resize the image",
		Params: []ParamMeta{
			{Name: "columns", Type: ParamTypeInt, Required: true, Min: float64Ptr(0), Hint: "Target width in pixels. Use 0 to keep aspect ratio if your UI supports that.", Example: "800", Unit: "px"},
			{Name: "rows", Type: ParamTypeInt, Required: true, Min: float64Ptr(0), Hint: "Target height in pixels. Use 0 to keep aspect ratio if your UI supports that.", Example: "600", Unit: "px"},
		},
	},
	{
		Name:        "adaptiveSharpen",
		Description: "Adaptively sharpen the image",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Size of the sharpening region in pixels. Lower = localized sharpening; higher = broader sharpening.", Example: "0.5"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Sharpen strength. Lower = subtle; higher = stronger (may introduce halos).", Example: "1.0"},
		},
	},
	{
		Name:        "adaptiveThreshold",
		Description: "Adaptively threshold the image",
		Params: []ParamMeta{
			{Name: "width", Type: ParamTypeInt, Required: true, Min: float64Ptr(1), Hint: "Block width in pixels used for local thresholding. Lower = finer local adaptation.", Example: "15", Unit: "px"},
			{Name: "height", Type: ParamTypeInt, Required: true, Min: float64Ptr(1), Hint: "Block height in pixels used for local thresholding. Lower = finer local adaptation.", Example: "15", Unit: "px"},
			{Name: "offset", Type: ParamTypeFloat, Required: true, Hint: "Offset applied during threshold test. Negative offsets favor black; positive favor white.", Example: "0.0"},
		},
	},
	{
		Name:        "addNoise",
		Description: "Add noise to the image",
		Params: []ParamMeta{
			{
				Name:        "noiseType",
				Type:        ParamTypeEnum,
				Required:    true,
				Hint:        "Choose the noise distribution to apply. Different types produce qualitatively different noise.",
				Example:     "GAUSSIAN",
				EnumOptions: []string{"UNDEFINED", "UNIFORM", "GAUSSIAN", "MULTIPLICATIVE", "IMPULSE", "LAPLACIAN", "POISSON", "RANDOM"},
			},
		},
	},
	{
		Name:        "annotate",
		Description: "Draw text onto the image",
		Params: []ParamMeta{
			{Name: "text", Type: ParamTypeString, Required: true, Hint: "Text to draw on the image.", Example: "Hello, World!"},
			{Name: "font", Type: ParamTypeString, Required: false, Hint: "Font family or path to a font file to use for text.", Example: "Arial"},
			{Name: "size", Type: ParamTypeFloat, Required: true, Min: float64Ptr(1.0), Hint: "Font size in points.", Example: "24.0"},
			{Name: "x", Type: ParamTypeInt, Required: true, Hint: "X coordinate for the text baseline origin.", Example: "10", Unit: "px"},
			{Name: "y", Type: ParamTypeInt, Required: true, Hint: "Y coordinate for the text baseline origin.", Example: "50", Unit: "px"},
			{Name: "color", Type: ParamTypeString, Required: true, Hint: "Text color (hex, rgb(), or name).", Example: "#ffffff"},
		},
	},
	{
		Name:        "autoGamma",
		Description: "Automatically adjust the image gamma",
		Params:      []ParamMeta{},
	},
	{
		Name:        "autoLevel",
		Description: "Automatically adjust the image levels",
		Params:      []ParamMeta{},
	},
	{
		Name:        "autoOrient",
		Description: "Automatically orient the image using EXIF Orientation",
		Params:      []ParamMeta{},
	},
	{
		Name:        "blackThreshold",
		Description: "Threshold the image to black and white using a black threshold color",
		Params: []ParamMeta{
			{Name: "threshold", Type: ParamTypeString, Required: true, Hint: "Color value (hex, rgb(), or name). Pixels darker or equal to this color become black.", Example: "#202020"},
		},
	},
	{
		Name:        "blueShift",
		Description: "Simulate a blue shift (increase blue channel influence)",
		Params: []ParamMeta{
			{Name: "factor", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Multiplier for blue shift. Lower ~ subtle; higher ~ stronger cool/blue cast.", Example: "1.0"},
		},
	},
	{
		Name:        "blur",
		Description: "Blur the image using a Gaussian convolution",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Blur kernel radius in pixels. 0 sometimes lets library auto-pick.", Example: "0.0"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Standard deviation (strength). Lower = subtle; higher = stronger blur.", Example: "1.5"},
		},
	},
	{
		Name:        "charcoal",
		Description: "Simulate a charcoal drawing",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Scale of charcoal effect; lower = finer strokes, higher = coarser strokes.", Example: "1.0"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Intensity/softening of strokes. Lower = crisper; higher = softer.", Example: "0.5"},
		},
	},
	{
		Name:        "colorize",
		Description: "Colorize (tint) the image with a given color and opacity",
		Params: []ParamMeta{
			{Name: "color", Type: ParamTypeString, Required: true, Hint: "Color value (hex, rgb(), or name) to apply as tint.", Example: "#ff0000"},
			{Name: "opacity", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Max: float64Ptr(1.0), Hint: "Opacity of the tint from 0.0 to 1.0.", Example: "0.5"},
		},
	},
	{
		Name:        "composite",
		Description: "Composite an image onto another",
		Params: []ParamMeta{
			{Name: "sourceImagePath", Type: ParamTypeString, Required: true, Hint: "Filesystem path or URL to the overlay/source image.", Example: "overlay.png"},
			{Name: "composeOperator", Type: ParamTypeEnum, Required: true, Hint: "Compositing operator / blend mode. Choose the desired blend behavior.", Example: "OVER", EnumOptions: []string{
				"UNDEFINED", "ALPHA", "ATOP", "BLEND", "BLUR", "BUMPMAP", "CHANGE_MASK", "CLEAR",
				"COLOR_BURN", "COLOR_DODGE", "COLORIZE", "COPY", "COPY_ALPHA", "COPY_BLACK", "COPY_BLUE",
				"COPY_CYAN", "COPY_GREEN", "COPY_MAGENTA", "COPY_RED", "COPY_YELLOW", "DARKEN",
				"DARKEN_INTENSITY", "DIFFERENCE", "DISPLACE", "DISSOLVE", "DISTORT", "DIVIDE__DST",
				"DIVIDE_SRC", "DST", "DST_ATOP", "DST_IN", "DST_OUT", "DST_OVER", "EXCLUSION",
				"HARD_LIGHT", "HARD_MIX", "HUE", "IN", "INTENSITY", "LIGHTEN", "LIGHTEN_INTENSITY",
				"LINEAR_BURN", "LINEAR_DODGE", "LINEAR_LIGHT", "LUMINIZE", "MATHEMATICS", "MINUS_DST",
				"MINUS_SRC", "MODULATE", "MODULUS_ADD", "MODULUS_SUBTRACT", "MULTIPLY", "NO", "OUT",
				"OVER", "OVERLAY", "PEGTOP_LIGHT", "PIN_LIGHT", "PLUS", "REPLACE", "SATURATE", "SCREEN",
				"SOFT_LIGHT", "SRC", "SRC_ATOP", "SRC_IN", "SRC_OUT", "SRC_OVER", "THRESHOLD", "VIVID_LIGHT",
				"XOR",
			}},
			{Name: "x", Type: ParamTypeInt, Required: true, Hint: "X offset in pixels where the source is placed relative to top-left.", Example: "100", Unit: "px"},
			{Name: "y", Type: ParamTypeInt, Required: true, Hint: "Y offset in pixels where the source is placed relative to top-left.", Example: "50", Unit: "px"},
		},
	},
	{
		Name:        "compress",
		Description: "Compress the image to reduce file size (lossy or lossless)",
		Params: []ParamMeta{
			{Name: "type", Type: ParamTypeEnum, Required: true, Hint: "Compression type to use. Choose based on desired balance of quality and size.", Example: "JPEG", EnumOptions: []string{"UNDEFINED", "NO", "BZIP", "DXT1", "DXT3", "DXT5", "FAX", "GROUP4", "JPEG", "JPEG2000", "LOSSLESS_JPEG", "LZW", "RLE", "ZIP", "ZIPS", "PIZ", "PXR24", "B44", "B44A", "LZMA", "JBIG1", "JBIG2"}},
			{Name: "quality", Type: ParamTypeInt, Required: true, Min: float64Ptr(1), Max: float64Ptr(100), Hint: "Quality level (1-100). Lower = smaller file size but more compression artifacts.", Example: "85"},
		},
	},
	{
		Name:        "contrast",
		Description: "Enhance or reduce the image contrast",
		Params: []ParamMeta{
			{Name: "sharpen", Type: ParamTypeBool, Required: true, Hint: "true = increase contrast (sharpen), false = decrease contrast (soften).", Example: "true"},
		},
	},
	{
		Name:        "contrastStretch",
		Description: "Stretch image contrast by remapping intensity range",
		Params: []ParamMeta{
			{Name: "low", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Max: float64Ptr(100.0), Hint: "Lower percent to clip (0-100).", Unit: "%", Example: "0.5"},
			{Name: "high", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Max: float64Ptr(100.0), Hint: "Upper percent to clip (0-100).", Unit: "%", Example: "99.5"},
		},
	},
	{
		Name:        "crop",
		Description: "Crop the image to a rectangle",
		Params: []ParamMeta{
			{Name: "width", Type: ParamTypeInt, Required: true, Min: float64Ptr(0), Hint: "Crop width in pixels.", Example: "800", Unit: "px"},
			{Name: "height", Type: ParamTypeInt, Required: true, Min: float64Ptr(0), Hint: "Crop height in pixels.", Example: "600", Unit: "px"},
			{Name: "x", Type: ParamTypeInt, Required: true, Hint: "X offset in pixels of the crop origin.", Example: "0", Unit: "px"},
			{Name: "y", Type: ParamTypeInt, Required: true, Hint: "Y offset in pixels of the crop origin.", Example: "0", Unit: "px"},
		},
	},
	{
		Name:        "deskew",
		Description: "Reduce skew in the image using an automatic algorithm",
		Params: []ParamMeta{
			{Name: "threshold", Type: ParamTypeFloat, Required: true, Hint: "Threshold used to detect skew; smaller values = more sensitive.", Example: "40.0"},
		},
	},
	{
		Name:        "despeckle",
		Description: "Reduce speckle noise in the image",
		Params:      []ParamMeta{},
	},
	{
		Name:        "edge",
		Description: "Detect edges in the image",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Filter radius for edge detection. Lower = detect thin details; higher = thicker edges.", Example: "1.0"},
		},
	},
	{
		Name:        "emboss",
		Description: "Create an embossed effect",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Neighborhood radius for embossing.", Example: "1.0"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Standard deviation controlling emboss strength.", Example: "0.5"},
		},
	},
	{
		Name:        "equalize",
		Description: "Equalize the image histogram to boost global contrast",
		Params:      []ParamMeta{},
	},
	{
		Name:        "enhance",
		Description: "Enhance image quality (reduce noise and improve clarity)",
		Params:      []ParamMeta{},
	},
	{
		Name:        "flip",
		Description: "Flip the image vertically (top ↔ bottom)",
		Params:      []ParamMeta{},
	},
	{
		Name:        "flop",
		Description: "Flip the image horizontally (left ↔ right)",
		Params:      []ParamMeta{},
	},
	{
		Name:        "gamma",
		Description: "Apply gamma correction",
		Params: []ParamMeta{
			{Name: "gamma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Gamma factor. < 1 brightens midtones; > 1 darkens midtones. 1.0 = neutral.", Example: "1.0"},
		},
	},
	{
		Name:        "grayscale",
		Description: "Convert the image to grayscale colorspace",
		Params:      []ParamMeta{},
	},
	{
		Name: "identify",
		Description: "Identify and display image metadata (format, dimensions, color depth, profiles, etc.)\n" +
			"This command does not modify the image; it only outputs information.",
		Params: []ParamMeta{},
	},
	{
		Name:        "level",
		Description: "Remap image levels (black point, gamma, white point)",
		Params: []ParamMeta{
			{Name: "blackPoint", Type: ParamTypeFloat, Required: true, Hint: "Black point (0-QuantumRange).", Example: "0.0"},
			{Name: "gamma", Type: ParamTypeFloat, Required: true, Hint: "Gamma adjustment value.", Example: "1.0"},
			{Name: "whitePoint", Type: ParamTypeFloat, Required: true, Hint: "White point (0-QuantumRange).", Example: "100.0"},
		},
	},
	{
		Name:        "medianFilter",
		Description: "Apply a median filter to reduce salt-and-pepper noise",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeInt, Required: true, Min: float64Ptr(0.0), Hint: "Radius for the median filter kernel.", Example: "1"},
		},
	},
	{
		Name:        "modulate",
		Description: "Adjust brightness, saturation and hue",
		Params: []ParamMeta{
			{Name: "brightness", Type: ParamTypeFloat, Required: true, Hint: "Brightness percent (100 = unchanged).", Example: "100.0"},
			{Name: "saturation", Type: ParamTypeFloat, Required: true, Hint: "Saturation percent (100 = unchanged).", Example: "100.0"},
			{Name: "hue", Type: ParamTypeFloat, Required: true, Hint: "Hue percent (100 = unchanged).", Example: "100.0"},
		},
	},
	{
		Name:        "monochrome",
		Description: "Convert the image to bilevel (pure black & white)",
		Params:      []ParamMeta{},
	},
	{
		Name:        "negate",
		Description: "Negate (invert) the colors of the image",
		Params: []ParamMeta{
			{Name: "only_gray", Type: ParamTypeBool, Required: true, Hint: "true = invert only grayscale channel; false = invert all channels (full negative).", Example: "false"},
		},
	},
	{
		Name:        "normalize",
		Description: "Normalize image to use full dynamic range",
		Params:      []ParamMeta{},
	},
	{
		Name:        "oilpaint",
		Description: "Simulate an oil painting effect",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Neighborhood radius in pixels. Lower = fine brush strokes; higher = broader strokes.", Example: "3.0"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Smoothness/intensity of the oil effect. Lower = more texture; higher = softer.", Example: "1.0"},
		},
	},
	{
		Name:        "posterize",
		Description: "Reduce the number of colors in the image (posterize)",
		Params: []ParamMeta{
			{Name: "levels", Type: ParamTypeInt, Required: true, Min: float64Ptr(1), Max: float64Ptr(256), Hint: "Number of color levels per channel. Lower = stronger posterization (fewer colors).", Example: "8"},
			{Name: "dither", Type: ParamTypeBool, Required: true, Hint: "Enable dithering to reduce visual banding (adds grain-like pattern).", Example: "true"},
		},
	},
	{
		Name:        "resize",
		Description: "Resize the image",
		Params: []ParamMeta{
			{Name: "width", Type: ParamTypeInt, Required: true, Min: float64Ptr(0), Hint: "Target width in pixels. Use 0 to preserve aspect ratio if supported by your UI.", Example: "1024", Unit: "px"},
			{Name: "height", Type: ParamTypeInt, Required: true, Min: float64Ptr(0), Hint: "Target height in pixels. Use 0 to preserve aspect ratio if supported by your UI.", Example: "768", Unit: "px"},
		},
	},
	{
		Name:        "rotate",
		Description: "Rotate the image",
		Params: []ParamMeta{
			{Name: "degrees", Type: ParamTypeFloat, Required: true, Hint: "Degrees to rotate. Positive values rotate clockwise (wraps beyond 360).", Example: "90.0", Unit: "deg"},
		},
	},
	{
		Name:        "sepia",
		Description: "Apply a sepia filter to the image",
		Params: []ParamMeta{
			{Name: "threshold", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Max: float64Ptr(100.0), Hint: "Strength/threshold for sepia toning. Lower = subtle; higher = stronger brown/yellow cast.", Example: "80"},
		},
	},
	{
		Name:        "sharpen",
		Description: "Sharpen the image",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Region size in pixels for sharpening. Lower = fine detail sharpening; higher = broader.", Example: "0.5"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Amount/strength of sharpening. Lower = subtle; higher = stronger (may produce halos).", Example: "1.0"},
		},
	},
	{
		Name:        "solarize",
		Description: "Solarize the image (partially invert pixels)",
		Params: []ParamMeta{
			{Name: "threshold", Type: ParamTypeFloat, Required: true, Hint: "Threshold at which pixels are inverted. Lower = subtle effect; higher = stronger inversion.", Example: "50.0"},
		},
	},
	{
		Name:        "strip",
		Description: "Remove image profiles and comments (strip metadata)",
		Params:      []ParamMeta{},
	},
	{
		Name:        "swirl",
		Description: "Swirl the image by a number of degrees",
		Params: []ParamMeta{
			{Name: "degrees", Type: ParamTypeFloat, Required: true, Hint: "Angle of swirl distortion. Lower = gentle; higher = dramatic twisting.", Example: "90.0", Unit: "deg"},
		},
	},
	{
		Name:        "threshold",
		Description: "Threshold the image to pure black and white",
		Params: []ParamMeta{
			{Name: "threshold", Type: ParamTypeFloat, Required: true, Hint: "Threshold value; pixels above become white, below become black.", Example: "128.0"},
		},
	},
	{
		Name:        "trim",
		Description: "Remove blank/background edges from the image",
		Params: []ParamMeta{
			{Name: "fuzz", Type: ParamTypePercent, Required: true, Min: float64Ptr(0.0), Max: float64Ptr(100.0), Hint: "Tolerance when matching border color. Lower = strict (only exact matches trimmed); higher = permissive (more aggressive trimming).", Example: "3.0", Unit: "%"},
		},
	},
	{
		Name:        "unsharp",
		Description: "Apply an unsharp mask to sharpen the image with advanced controls",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Unit: "px", Hint: "Radius of the unsharp mask.", Example: "1.0"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Standard deviation for the Gaussian blur used by the mask.", Example: "0.5"},
			{Name: "amount", Type: ParamTypeFloat, Required: true, Hint: "Sharpening amount (multiplier).", Example: "1.5"},
			{Name: "threshold", Type: ParamTypeFloat, Required: true, Hint: "Threshold to limit sharpening to significant edges.", Example: "0.05"},
		},
	},
	{
		Name:        "vignette",
		Description: "Apply a vignette effect to darken or tint edges",
		Params: []ParamMeta{
			{Name: "radius", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Radius of the vignette effect.", Example: "50.0"},
			{Name: "sigma", Type: ParamTypeFloat, Required: true, Min: float64Ptr(0.0), Hint: "Feathering/smoothing of the vignette.", Example: "20.0"},
			{Name: "x", Type: ParamTypeInt, Required: true, Hint: "X coordinate of the vignette center.", Example: "0", Unit: "px"},
			{Name: "y", Type: ParamTypeInt, Required: true, Hint: "Y coordinate of the vignette center.", Example: "0", Unit: "px"},
		},
	},
}
