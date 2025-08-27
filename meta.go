package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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

// mapEnumToNumeric attempts to translate some known enum textual values to numeric IDs
// expected by ApplyCommand. Extend these maps as needed.
func mapEnumToNumeric(paramName string, val string) (string, bool) {
	v := strings.TrimSpace(val)
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		// already numeric
		return v, true
	}

	noiseTypeMap := map[string]int64{
		"UNIFORM":        0,
		"GAUSSIAN":       1,
		"MULTIPLICATIVE": 2,
		"IMPULSE":        3,
		"LAPLACIAN":      4,
		"POISSON":        5,
	}

	composeOpMap := map[string]int64{
		"OVER":     0,
		"IN":       1,
		"OUT":      2,
		"ATOP":     3,
		"XOR":      4,
		"MULTIPLY": 5,
		"SCREEN":   6,
		"ADD":      7,
		"SUBTRACT": 8,
	}

	switch strings.ToLower(paramName) {
	case "noisetype", "noise_type", "noise":
		if id, ok := noiseTypeMap[strings.ToUpper(v)]; ok {
			return strconv.FormatInt(id, 10), true
		}
	case "composeoperator", "compose_operator", "compose":
		if id, ok := composeOpMap[strings.ToUpper(v)]; ok {
			return strconv.FormatInt(id, 10), true
		}
	}

	// Not a known mapping
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
