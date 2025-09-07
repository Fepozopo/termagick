package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// ParamType is a small enum for parameter types used in metadata.
type ParamType string

const (
	ParamTypeInt     ParamType = "int"
	ParamTypeFloat   ParamType = "float"
	ParamTypeBool    ParamType = "bool"
	ParamTypeString  ParamType = "string"
	ParamTypeEnum    ParamType = "enum"
	ParamTypePercent ParamType = "percent"
)

// ParamMeta describes a single parameter for a command.
type ParamMeta struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Required    bool      `json:"required"`
	Min         *float64  `json:"min,omitempty"`
	Max         *float64  `json:"max,omitempty"`
	Unit        string    `json:"unit,omitempty"`
	Hint        string    `json:"hint,omitempty"`
	Example     string    `json:"example,omitempty"`
	EnumOptions []string  `json:"enumOptions,omitempty"`
}

// CommandMeta ties a command name to its params and description.
type CommandMeta struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Params      []ParamMeta `json:"params"`
}

// ValidationRule is a machine-friendly representation of the constraints
// that a UI or client can use to validate input before invoking a command.
type ValidationRule struct {
	Type        ParamType `json:"type"`
	Required    bool      `json:"required"`
	Min         *float64  `json:"min,omitempty"`
	Max         *float64  `json:"max,omitempty"`
	Unit        string    `json:"unit,omitempty"`
	Pattern     string    `json:"pattern,omitempty"`     // optional regex-like pattern or note
	EnumOptions []string  `json:"enumOptions,omitempty"` // valid when Type == ParamTypeEnum
	Example     string    `json:"example,omitempty"`
	Hint        string    `json:"hint,omitempty"`
}

// LoadCommandMetaFromFile reads a JSON file containing []CommandMeta and unmarshals it.
func LoadCommandMetaFromFile(path string) ([]CommandMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read metadata file: %w", err)
	}
	var meta []CommandMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal metadata json: %w", err)
	}
	return meta, nil
}

// GetCommandMetaByName looks up a CommandMeta by name from a slice.
// Returns nil if not found.
func GetCommandMetaByName(all []CommandMeta, name string) *CommandMeta {
	for i := range all {
		if all[i].Name == name {
			return &all[i]
		}
	}
	return nil
}

// GenerateTooltip produces a human-friendly tooltip string for the command.
// The output is intended for UI tooltips/help text.
func GenerateTooltip(cmd CommandMeta) string {
	var sb strings.Builder
	sb.WriteString(cmd.Description)
	if len(cmd.Params) == 0 {
		sb.WriteString(" — no parameters")
		return sb.String()
	}
	sb.WriteString(" — parameters:\n")
	for _, p := range cmd.Params {
		req := "optional"
		if p.Required {
			req = "required"
		}
		typeLabel := string(p.Type)
		if p.Type == ParamTypeEnum && len(p.EnumOptions) > 0 {
			typeLabel = fmt.Sprintf("enum(%s)", strings.Join(p.EnumOptions, "|"))
		}
		line := fmt.Sprintf("- %s (%s, %s)", p.Name, typeLabel, req)
		sb.WriteString(line)
		parts := []string{}
		if p.Unit != "" {
			parts = append(parts, fmt.Sprintf("unit: %s", p.Unit))
		}
		if p.Hint != "" {
			parts = append(parts, p.Hint)
		}
		if p.Example != "" {
			parts = append(parts, fmt.Sprintf("example: %s", p.Example))
		}
		if len(parts) > 0 {
			sb.WriteString(" — " + strings.Join(parts, "; "))
		}
		sb.WriteString("\n")
	}
	return strings.TrimSpace(sb.String())
}

// GenerateValidationRules returns a map keyed by parameter name that describes
// validation constraints and UI control hints for each parameter.
func GenerateValidationRules(cmd CommandMeta) map[string]ValidationRule {
	rules := make(map[string]ValidationRule, len(cmd.Params))
	for _, p := range cmd.Params {
		r := ValidationRule{
			Type:        p.Type,
			Required:    p.Required,
			Min:         p.Min,
			Max:         p.Max,
			Unit:        p.Unit,
			EnumOptions: p.EnumOptions,
			Example:     p.Example,
			Hint:        p.Hint,
		}
		rules[p.Name] = r
	}
	return rules
}

// MetaStore is a lightweight in-memory store for command metadata.
type MetaStore struct {
	Commands []CommandMeta
	byName   map[string]CommandMeta
}

// NewMetaStoreFromFile creates a MetaStore by reading metadata from a JSON file.
func NewMetaStoreFromFile(path string) (*MetaStore, error) {
	cmds, err := LoadCommandMetaFromFile(path)
	if err != nil {
		return nil, err
	}
	return NewMetaStore(cmds), nil
}

// NewMetaStore creates a MetaStore from an in-memory slice.
func NewMetaStore(cmds []CommandMeta) *MetaStore {
	ms := &MetaStore{Commands: cmds, byName: make(map[string]CommandMeta, len(cmds))}
	for _, c := range cmds {
		ms.byName[c.Name] = c
	}
	return ms
}

// GetTooltip returns the tooltip string for the named command.
func (m *MetaStore) GetTooltip(name string) (string, error) {
	c, ok := m.byName[name]
	if !ok {
		return "", fmt.Errorf("unknown command: %s", name)
	}
	return GenerateTooltip(c), nil
}

// GetValidationRules returns the validation rules for the named command.
func (m *MetaStore) GetValidationRules(name string) (map[string]ValidationRule, error) {
	c, ok := m.byName[name]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", name)
	}
	return GenerateValidationRules(c), nil
}

// GetCommandHelp returns both a human-friendly tooltip string and a machine-friendly
// validation rules map for the named command.
func (m *MetaStore) GetCommandHelp(name string) (string, map[string]ValidationRule, error) {
	c, ok := m.byName[name]
	if !ok {
		return "", nil, fmt.Errorf("unknown command: %s", name)
	}
	tooltip := GenerateTooltip(c)
	rules := GenerateValidationRules(c)
	return tooltip, rules, nil
}

// Helper to create float64 pointer for convenience when building metadata in code.
func float64Ptr(v float64) *float64 { return &v }

// parseBoolLike accepts common truthy/falsy forms and returns "true"/"false" string.
func parseBoolLikeToString(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "t", "true", "y", "yes", "on":
		return "true", nil
	case "0", "f", "false", "n", "no", "off":
		return "false", nil
	default:
		return "", fmt.Errorf("invalid boolean: %q", s)
	}
}

// parsePercentValue parses a percent string like "3%" or a bare number and returns numeric string.
func parsePercentValue(s string) (string, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		raw := strings.TrimSuffix(s, "%")
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return "", fmt.Errorf("invalid percent value: %q", s)
		}
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	}
	// bare number
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return "", fmt.Errorf("invalid percent/float value: %q", s)
	}
	return s, nil
}

/*
Package-level enum maps and helpers.

We extract the textual<->numeric maps into package-level variables so that
they can be reused by both directions of mapping:

- mapEnumToNumeric: textual -> numeric (existing behavior)
- mapNumericToEnumName: numeric -> textual (new helper used elsewhere, e.g. GetImageInfo)

Add or extend maps here as new enum types are required.
*/

var (
	noiseTypeNameToValue = map[string]int64{
		"UNDEFINED":      int64(imagick.NOISE_UNDEFINED),
		"UNIFORM":        int64(imagick.NOISE_UNIFORM),
		"GAUSSIAN":       int64(imagick.NOISE_GAUSSIAN),
		"MULTIPLICATIVE": int64(imagick.NOISE_MULTIPLICATIVE_GAUSSIAN),
		"IMPULSE":        int64(imagick.NOISE_IMPULSE),
		"LAPLACIAN":      int64(imagick.NOISE_LAPLACIAN),
		"POISSON":        int64(imagick.NOISE_POISSON),
		"RANDOM":         int64(imagick.NOISE_RANDOM),
	}

	noiseTypeValueToName = map[int64]string{
		int64(imagick.NOISE_UNDEFINED):               "UNDEFINED",
		int64(imagick.NOISE_UNIFORM):                 "UNIFORM",
		int64(imagick.NOISE_GAUSSIAN):                "GAUSSIAN",
		int64(imagick.NOISE_MULTIPLICATIVE_GAUSSIAN): "MULTIPLICATIVE",
		int64(imagick.NOISE_IMPULSE):                 "IMPULSE",
		int64(imagick.NOISE_LAPLACIAN):               "LAPLACIAN",
		int64(imagick.NOISE_POISSON):                 "POISSON",
		int64(imagick.NOISE_RANDOM):                  "RANDOM",
	}

	// Compose operator textual aliases mapped to ImageMagick composite operator constants.
	composeOpNameToValue = map[string]int64{
		"UNDEFINED":         int64(imagick.COMPOSITE_OP_UNDEFINED),
		"ALPHA":             int64(imagick.COMPOSITE_OP_ALPHA),
		"ATOP":              int64(imagick.COMPOSITE_OP_ATOP),
		"BLEND":             int64(imagick.COMPOSITE_OP_BLEND),
		"BLUR":              int64(imagick.COMPOSITE_OP_BLUR),
		"BUMPMAP":           int64(imagick.COMPOSITE_OP_BUMPMAP),
		"CHANGE_MASK":       int64(imagick.COMPOSITE_OP_CHANGE_MASK),
		"CLEAR":             int64(imagick.COMPOSITE_OP_CLEAR),
		"COLOR_BURN":        int64(imagick.COMPOSITE_OP_COLOR_BURN),
		"COLOR_DODGE":       int64(imagick.COMPOSITE_OP_COLOR_DODGE),
		"COLORIZE":          int64(imagick.COMPOSITE_OP_COLORIZE),
		"COPY":              int64(imagick.COMPOSITE_OP_COPY),
		"COPY_ALPHA":        int64(imagick.COMPOSITE_OP_COPY_ALPHA),
		"COPY_BLACK":        int64(imagick.COMPOSITE_OP_COPY_BLACK),
		"COPY_BLUE":         int64(imagick.COMPOSITE_OP_COPY_BLUE),
		"COPY_CYAN":         int64(imagick.COMPOSITE_OP_COPY_CYAN),
		"COPY_GREEN":        int64(imagick.COMPOSITE_OP_COPY_GREEN),
		"COPY_MAGENTA":      int64(imagick.COMPOSITE_OP_COPY_MAGENTA),
		"COPY_RED":          int64(imagick.COMPOSITE_OP_COPY_RED),
		"COPY_YELLOW":       int64(imagick.COMPOSITE_OP_COPY_YELLOW),
		"DARKEN":            int64(imagick.COMPOSITE_OP_DARKEN),
		"DARKEN_INTENSITY":  int64(imagick.COMPOSITE_OP_DARKEN_INTENSITY),
		"DIFFERENCE":        int64(imagick.COMPOSITE_OP_DIFFERENCE),
		"DISPLACE":          int64(imagick.COMPOSITE_OP_DISPLACE),
		"DISSOLVE":          int64(imagick.COMPOSITE_OP_DISSOLVE),
		"DISTORT":           int64(imagick.COMPOSITE_OP_DISTORT),
		"DIVIDE__DST":       int64(imagick.COMPOSITE_OP_DIVIDE__DST),
		"DIVIDE_SRC":        int64(imagick.COMPOSITE_OP_DIVIDE_SRC),
		"DST":               int64(imagick.COMPOSITE_OP_DST),
		"DST_ATOP":          int64(imagick.COMPOSITE_OP_DST_ATOP),
		"DST_IN":            int64(imagick.COMPOSITE_OP_DST_IN),
		"DST_OUT":           int64(imagick.COMPOSITE_OP_DST_OUT),
		"DST_OVER":          int64(imagick.COMPOSITE_OP_DST_OVER),
		"EXCLUSION":         int64(imagick.COMPOSITE_OP_EXCLUSION),
		"HARD_LIGHT":        int64(imagick.COMPOSITE_OP_HARD_LIGHT),
		"HARD_MIX":          int64(imagick.COMPOSITE_OP_HARD_MIX),
		"HUE":               int64(imagick.COMPOSITE_OP_HUE),
		"IN":                int64(imagick.COMPOSITE_OP_IN),
		"INTENSITY":         int64(imagick.COMPOSITE_OP_INTENSITY),
		"LIGHTEN":           int64(imagick.COMPOSITE_OP_LIGHTEN),
		"LIGHTEN_INTENSITY": int64(imagick.COMPOSITE_OP_LIGHTEN_INTENSITY),
		"LINEAR_BURN":       int64(imagick.COMPOSITE_OP_LINEAR_BURN),
		"LINEAR_DODGE":      int64(imagick.COMPOSITE_OP_LINEAR_DODGE),
		"LINEAR_LIGHT":      int64(imagick.COMPOSITE_OP_LINEAR_LIGHT),
		"LUMINIZE":          int64(imagick.COMPOSITE_OP_LUMINIZE),
		"MATHEMATICS":       int64(imagick.COMPOSITE_OP_MATHEMATICS),
		"MINUS_DST":         int64(imagick.COMPOSITE_OP_MINUS_DST),
		"MINUS_SRC":         int64(imagick.COMPOSITE_OP_MINUS_SRC),
		"MODULATE":          int64(imagick.COMPOSITE_OP_MODULATE),
		"MODULUS_ADD":       int64(imagick.COMPOSITE_OP_MODULUS_ADD),
		"MODULUS_SUBTRACT":  int64(imagick.COMPOSITE_OP_MODULUS_SUBTRACT),
		"MULTIPLY":          int64(imagick.COMPOSITE_OP_MULTIPLY),
		"NO":                int64(imagick.COMPOSITE_OP_NO),
		"OUT":               int64(imagick.COMPOSITE_OP_OUT),
		"OVER":              int64(imagick.COMPOSITE_OP_OVER),
		"OVERLAY":           int64(imagick.COMPOSITE_OP_OVERLAY),
		"PEGTOP_LIGHT":      int64(imagick.COMPOSITE_OP_PEGTOP_LIGHT),
		"PIN_LIGHT":         int64(imagick.COMPOSITE_OP_PIN_LIGHT),
		"PLUS":              int64(imagick.COMPOSITE_OP_PLUS),
		"REPLACE":           int64(imagick.COMPOSITE_OP_REPLACE),
		"SATURATE":          int64(imagick.COMPOSITE_OP_SATURATE),
		"SCREEN":            int64(imagick.COMPOSITE_OP_SCREEN),
		"SOFT_LIGHT":        int64(imagick.COMPOSITE_OP_SOFT_LIGHT),
		"SRC":               int64(imagick.COMPOSITE_OP_SRC),
		"SRC_ATOP":          int64(imagick.COMPOSITE_OP_SRC_ATOP),
		"SRC_IN":            int64(imagick.COMPOSITE_OP_SRC_IN),
		"SRC_OUT":           int64(imagick.COMPOSITE_OP_SRC_OUT),
		"SRC_OVER":          int64(imagick.COMPOSITE_OP_SRC_OVER),
		"THRESHOLD":         int64(imagick.COMPOSITE_OP_THRESHOLD),
		"VIVID_LIGHT":       int64(imagick.COMPOSITE_OP_VIVID_LIGHT),
		"XOR":               int64(imagick.COMPOSITE_OP_XOR),
	}

	composeOpValueToName = map[int64]string{
		int64(imagick.COMPOSITE_OP_UNDEFINED):         "UNDEFINED",
		int64(imagick.COMPOSITE_OP_ALPHA):             "ALPHA",
		int64(imagick.COMPOSITE_OP_ATOP):              "ATOP",
		int64(imagick.COMPOSITE_OP_BLEND):             "BLEND",
		int64(imagick.COMPOSITE_OP_BLUR):              "BLUR",
		int64(imagick.COMPOSITE_OP_BUMPMAP):           "BUMPMAP",
		int64(imagick.COMPOSITE_OP_CHANGE_MASK):       "CHANGE_MASK",
		int64(imagick.COMPOSITE_OP_CLEAR):             "CLEAR",
		int64(imagick.COMPOSITE_OP_COLOR_BURN):        "COLOR_BURN",
		int64(imagick.COMPOSITE_OP_COLOR_DODGE):       "COLOR_DODGE",
		int64(imagick.COMPOSITE_OP_COLORIZE):          "COLORIZE",
		int64(imagick.COMPOSITE_OP_COPY):              "COPY",
		int64(imagick.COMPOSITE_OP_COPY_ALPHA):        "COPY_ALPHA",
		int64(imagick.COMPOSITE_OP_COPY_BLACK):        "COPY_BLACK",
		int64(imagick.COMPOSITE_OP_COPY_BLUE):         "COPY_BLUE",
		int64(imagick.COMPOSITE_OP_COPY_CYAN):         "COPY_CYAN",
		int64(imagick.COMPOSITE_OP_COPY_GREEN):        "COPY_GREEN",
		int64(imagick.COMPOSITE_OP_COPY_MAGENTA):      "COPY_MAGENTA",
		int64(imagick.COMPOSITE_OP_COPY_RED):          "COPY_RED",
		int64(imagick.COMPOSITE_OP_COPY_YELLOW):       "COPY_YELLOW",
		int64(imagick.COMPOSITE_OP_DARKEN):            "DARKEN",
		int64(imagick.COMPOSITE_OP_DARKEN_INTENSITY):  "DARKEN_INTENSITY",
		int64(imagick.COMPOSITE_OP_DIFFERENCE):        "DIFFERENCE",
		int64(imagick.COMPOSITE_OP_DISPLACE):          "DISPLACE",
		int64(imagick.COMPOSITE_OP_DISSOLVE):          "DISSOLVE",
		int64(imagick.COMPOSITE_OP_DISTORT):           "DISTORT",
		int64(imagick.COMPOSITE_OP_DIVIDE__DST):       "DIVIDE__DST",
		int64(imagick.COMPOSITE_OP_DIVIDE_SRC):        "DIVIDE_SRC",
		int64(imagick.COMPOSITE_OP_DST):               "DST",
		int64(imagick.COMPOSITE_OP_DST_ATOP):          "DST_ATOP",
		int64(imagick.COMPOSITE_OP_DST_IN):            "DST_IN",
		int64(imagick.COMPOSITE_OP_DST_OUT):           "DST_OUT",
		int64(imagick.COMPOSITE_OP_DST_OVER):          "DST_OVER",
		int64(imagick.COMPOSITE_OP_EXCLUSION):         "EXCLUSION",
		int64(imagick.COMPOSITE_OP_HARD_LIGHT):        "HARD_LIGHT",
		int64(imagick.COMPOSITE_OP_HARD_MIX):          "HARD_MIX",
		int64(imagick.COMPOSITE_OP_HUE):               "HUE",
		int64(imagick.COMPOSITE_OP_IN):                "IN",
		int64(imagick.COMPOSITE_OP_INTENSITY):         "INTENSITY",
		int64(imagick.COMPOSITE_OP_LIGHTEN):           "LIGHTEN",
		int64(imagick.COMPOSITE_OP_LIGHTEN_INTENSITY): "LIGHTEN_INTENSITY",
		int64(imagick.COMPOSITE_OP_LINEAR_BURN):       "LINEAR_BURN",
		int64(imagick.COMPOSITE_OP_LINEAR_DODGE):      "LINEAR_DODGE",
		int64(imagick.COMPOSITE_OP_LINEAR_LIGHT):      "LINEAR_LIGHT",
		int64(imagick.COMPOSITE_OP_LUMINIZE):          "LUMINIZE",
		int64(imagick.COMPOSITE_OP_MATHEMATICS):       "MATHEMATICS",
		int64(imagick.COMPOSITE_OP_MINUS_DST):         "MINUS_DST",
		int64(imagick.COMPOSITE_OP_MINUS_SRC):         "MINUS_SRC",
		int64(imagick.COMPOSITE_OP_MODULATE):          "MODULATE",
		int64(imagick.COMPOSITE_OP_MODULUS_ADD):       "MODULUS_ADD",
		int64(imagick.COMPOSITE_OP_MODULUS_SUBTRACT):  "MODULUS_SUBTRACT",
		int64(imagick.COMPOSITE_OP_MULTIPLY):          "MULTIPLY",
		int64(imagick.COMPOSITE_OP_NO):                "NO",
		int64(imagick.COMPOSITE_OP_OUT):               "OUT",
		int64(imagick.COMPOSITE_OP_OVER):              "OVER",
		int64(imagick.COMPOSITE_OP_OVERLAY):           "OVERLAY",
		int64(imagick.COMPOSITE_OP_PEGTOP_LIGHT):      "PEGTOP_LIGHT",
		int64(imagick.COMPOSITE_OP_PIN_LIGHT):         "PIN_LIGHT",
		int64(imagick.COMPOSITE_OP_PLUS):              "PLUS",
		int64(imagick.COMPOSITE_OP_REPLACE):           "REPLACE",
		int64(imagick.COMPOSITE_OP_SATURATE):          "SATURATE",
		int64(imagick.COMPOSITE_OP_SCREEN):            "SCREEN",
		int64(imagick.COMPOSITE_OP_SOFT_LIGHT):        "SOFT_LIGHT",
		int64(imagick.COMPOSITE_OP_SRC):               "SRC",
		int64(imagick.COMPOSITE_OP_SRC_ATOP):          "SRC_ATOP",
		int64(imagick.COMPOSITE_OP_SRC_IN):            "SRC_IN",
		int64(imagick.COMPOSITE_OP_SRC_OUT):           "SRC_OUT",
		int64(imagick.COMPOSITE_OP_SRC_OVER):          "SRC_OVER",
		int64(imagick.COMPOSITE_OP_THRESHOLD):         "THRESHOLD",
		int64(imagick.COMPOSITE_OP_VIVID_LIGHT):       "VIVID_LIGHT",
		int64(imagick.COMPOSITE_OP_XOR):               "XOR",
	}

	// Compression textual aliases mapped to ImageMagick compression constants.
	// This allows UI-level enum normalization to produce the exact numeric IDs
	// that ApplyCommand expects (we use the imagick package constants).
	compressionNameToValue = map[string]int64{
		"UNDEFINED":     int64(imagick.COMPRESSION_UNDEFINED),
		"NO":            int64(imagick.COMPRESSION_NO),
		"BZIP":          int64(imagick.COMPRESSION_BZIP),
		"DXT1":          int64(imagick.COMPRESSION_DXT1),
		"DXT3":          int64(imagick.COMPRESSION_DXT3),
		"DXT5":          int64(imagick.COMPRESSION_DXT5),
		"FAX":           int64(imagick.COMPRESSION_FAX),
		"GROUP4":        int64(imagick.COMPRESSION_GROUP4),
		"JPEG":          int64(imagick.COMPRESSION_JPEG),
		"JPEG2000":      int64(imagick.COMPRESSION_JPEG2000),
		"LOSSLESS_JPEG": int64(imagick.COMPRESSION_LOSSLESS_JPEG),
		"LZW":           int64(imagick.COMPRESSION_LZW),
		"RLE":           int64(imagick.COMPRESSION_RLE),
		"ZIP":           int64(imagick.COMPRESSION_ZIP),
		"ZIPS":          int64(imagick.COMPRESSION_ZIPS),
		"PIZ":           int64(imagick.COMPRESSION_PIZ),
		"PXR24":         int64(imagick.COMPRESSION_PXR24),
		"B44":           int64(imagick.COMPRESSION_B44),
		"B44A":          int64(imagick.COMPRESSION_B44A),
		"LZMA":          int64(imagick.COMPRESSION_LZMA),
		"JBIG1":         int64(imagick.COMPRESSION_JBIG1),
		"JBIG2":         int64(imagick.COMPRESSION_JBIG2),
	}

	compressionValueToName = map[int64]string{
		int64(imagick.COMPRESSION_UNDEFINED):     "UNDEFINED",
		int64(imagick.COMPRESSION_NO):            "NO",
		int64(imagick.COMPRESSION_BZIP):          "BZIP",
		int64(imagick.COMPRESSION_DXT1):          "DXT1",
		int64(imagick.COMPRESSION_DXT3):          "DXT3",
		int64(imagick.COMPRESSION_DXT5):          "DXT5",
		int64(imagick.COMPRESSION_FAX):           "FAX",
		int64(imagick.COMPRESSION_GROUP4):        "GROUP4",
		int64(imagick.COMPRESSION_JPEG):          "JPEG",
		int64(imagick.COMPRESSION_JPEG2000):      "JPEG2000",
		int64(imagick.COMPRESSION_LOSSLESS_JPEG): "LOSSLESS_JPEG",
		int64(imagick.COMPRESSION_LZW):           "LZW",
		int64(imagick.COMPRESSION_RLE):           "RLE",
		int64(imagick.COMPRESSION_ZIP):           "ZIP",
		int64(imagick.COMPRESSION_ZIPS):          "ZIPS",
		int64(imagick.COMPRESSION_PIZ):           "PIZ",
		int64(imagick.COMPRESSION_PXR24):         "PXR24",
		int64(imagick.COMPRESSION_B44):           "B44",
		int64(imagick.COMPRESSION_B44A):          "B44A",
		int64(imagick.COMPRESSION_LZMA):          "LZMA",
		int64(imagick.COMPRESSION_JBIG1):         "JBIG1",
		int64(imagick.COMPRESSION_JBIG2):         "JBIG2",
	}
)

// mapEnumToNumeric attempts to translate some known enum textual values to numeric IDs
// expected by ApplyCommand. Extend these maps as needed.
func mapEnumToNumeric(paramName string, val string) (string, bool) {
	v := strings.TrimSpace(val)
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		// already numeric
		return v, true
	}

	switch strings.ToLower(paramName) {
	case "noisetype", "noise_type", "noise":
		if id, ok := noiseTypeNameToValue[strings.ToUpper(v)]; ok {
			return strconv.FormatInt(id, 10), true
		}
	case "composeoperator", "compose_operator", "compose":
		if id, ok := composeOpNameToValue[strings.ToUpper(v)]; ok {
			return strconv.FormatInt(id, 10), true
		}
	case "type", "compression", "compressiontype", "compress":
		if id, ok := compressionNameToValue[strings.ToUpper(v)]; ok {
			return strconv.FormatInt(id, 10), true
		}
	}

	// Not a known mapping
	return "", false
}

// mapNumericToEnumName attempts the reverse mapping: given a parameter name and
// an integer value, return the canonical textual name (if known).
// This is useful when you have numeric enum values (e.g. from imagick) and want
// to render or report the textual alias.
func mapNumericToEnumName(paramName string, id int64) (string, bool) {
	switch strings.ToLower(paramName) {
	case "noisetype", "noise_type", "noise":
		if s, ok := noiseTypeValueToName[id]; ok {
			return s, true
		}
	case "composeoperator", "compose_operator", "compose":
		if s, ok := composeOpValueToName[id]; ok {
			return s, true
		}
	case "type", "compression", "compressiontype", "compress":
		if s, ok := compressionValueToName[id]; ok {
			return s, true
		}
	}
	return "", false
}

// NormalizeArgs normalizes and validates the provided args (user-provided strings)
// for the given command name using metadata in the provided MetaStore.
//
// The function performs:
//   - required param presence checks
//   - boolean normalization (accepts yes/no/1/0 etc. -> "true"/"false")
//   - percent parsing (e.g., "3%" -> "3")
//   - enum textual -> numeric mapping for known enums (noiseType, composeOperator)
//   - basic range checking using Min/Max present in metadata
//
// Returns a new slice of args (same length as command params) suitable for passing
// directly to ApplyCommand (which expects string representations the existing code parses).
func NormalizeArgs(store *MetaStore, cmdName string, args []string) ([]string, error) {
	if store == nil {
		return nil, fmt.Errorf("metadata store is nil")
	}
	cmdMeta, ok := store.byName[cmdName]
	if !ok {
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}

	out := make([]string, len(cmdMeta.Params))

	for i, p := range cmdMeta.Params {
		var raw string
		if i < len(args) {
			raw = strings.TrimSpace(args[i])
		} else {
			raw = ""
		}

		// Required check
		if raw == "" {
			if p.Required {
				return nil, fmt.Errorf("missing required parameter: %s", p.Name)
			}
			out[i] = ""
			continue
		}

		switch p.Type {
		case ParamTypeInt:
			// ensure integer and range
			v, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parameter %s: expected integer, got %q", p.Name, raw)
			}
			if p.Min != nil && float64(v) < *p.Min {
				return nil, fmt.Errorf("parameter %s: %d < min %v", p.Name, v, *p.Min)
			}
			if p.Max != nil && float64(v) > *p.Max {
				return nil, fmt.Errorf("parameter %s: %d > max %v", p.Name, v, *p.Max)
			}
			out[i] = strconv.FormatInt(v, 10)

		case ParamTypeFloat:
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return nil, fmt.Errorf("parameter %s: expected float, got %q", p.Name, raw)
			}
			if p.Min != nil && f < *p.Min {
				return nil, fmt.Errorf("parameter %s: %v < min %v", p.Name, f, *p.Min)
			}
			if p.Max != nil && f > *p.Max {
				return nil, fmt.Errorf("parameter %s: %v > max %v", p.Name, f, *p.Max)
			}
			out[i] = strconv.FormatFloat(f, 'f', -1, 64)

		case ParamTypePercent:
			// allow "3%" or "3" and return numeric form (no %)
			n, err := parsePercentValue(raw)
			if err != nil {
				return nil, fmt.Errorf("parameter %s: %w", p.Name, err)
			}
			// optional range enforcement
			f, _ := strconv.ParseFloat(n, 64)
			if p.Min != nil && f < *p.Min {
				return nil, fmt.Errorf("parameter %s: %v < min %v", p.Name, f, *p.Min)
			}
			if p.Max != nil && f > *p.Max {
				return nil, fmt.Errorf("parameter %s: %v > max %v", p.Name, f, *p.Max)
			}
			out[i] = n

		case ParamTypeBool:
			bs, err := parseBoolLikeToString(raw)
			if err != nil {
				return nil, fmt.Errorf("parameter %s: %w", p.Name, err)
			}
			out[i] = bs

		case ParamTypeEnum:
			// Try numeric first
			if _, err := strconv.ParseInt(raw, 10, 64); err == nil {
				out[i] = raw
				break
			}
			// Known mappings (noiseType, composeOperator, etc.)
			if mapped, ok := mapEnumToNumeric(p.Name, raw); ok {
				out[i] = mapped
				break
			}
			// If the metadata provides EnumOptions, try to resolve to index as fallback.
			if len(p.EnumOptions) > 0 {
				found := -1
				for idx, opt := range p.EnumOptions {
					if strings.EqualFold(opt, raw) {
						found = idx
						break
					}
				}
				if found >= 0 {
					// NOTE: this fallback returns the zero-based index of the option.
					// This may not match ImageMagick's constant values for the enum, but
					// is provided as a best-effort fallback. Prefer adding explicit maps
					// above for enums that must match specific C constants.
					out[i] = strconv.Itoa(found)
					break
				}
			}
			// Give the user a helpful error listing allowed options
			if len(p.EnumOptions) > 0 {
				return nil, fmt.Errorf("parameter %s: unknown option %q, allowed: %v", p.Name, raw, p.EnumOptions)
			}
			return nil, fmt.Errorf("parameter %s: cannot map enum value %q to numeric form", p.Name, raw)

		case ParamTypeString:
			out[i] = raw

		default:
			return nil, fmt.Errorf("parameter %s: unsupported param type %q", p.Name, p.Type)
		}
	}

	return out, nil
}
