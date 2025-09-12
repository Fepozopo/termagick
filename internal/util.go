package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// PromptLine displays a prompt and reads a full line of input from the user.
// The returned string is trimmed of surrounding whitespace (including the newline).
func PromptLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// PromptLineOrFzf reads a full line from stdin and treats a single-line "/"
// as a request to invoke fzf for file selection. Behavior:
//   - Print the prompt.
//   - Read a full line (including spaces).
//   - If the trimmed line equals "/", launch fzf via SelectFileWithFzf(".").
//   - If fzf returns a non-empty selection, return it.
//   - If fzf is unavailable or selection is cancelled, fall back to a typed prompt
//     (re-using PromptLine to read a full line).
//   - Otherwise return the trimmed line as the input value.
//
// This approach preserves support for paths containing spaces because we read
// the entire input line instead of a single token.
func PromptLineOrFzf(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input := strings.TrimSpace(line)

	if input == "/" {
		// User requested fzf selection.
		sel, selErr := SelectFileWithFzf(".")
		if selErr == nil && sel != "" {
			// Show concise indicator and return the selection.
			fmt.Printf(" [fzf] %s\n", sel)
			return sel, nil
		}
		// fzf not available or selection cancelled — fall back to typed prompt.
		return PromptLine(prompt)
	}

	return input, nil
}

// PromptLineWithFzfReader is a convenience variant that reads from the provided
// bufio.Reader. This is useful when the caller already has a reader instance
// and wants to avoid creating a new one (ensures no input is lost to a
// separate buffered reader).
func PromptLineWithFzfReader(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input := strings.TrimSpace(line)

	if input == "/" {
		sel, selErr := SelectFileWithFzf(".")
		if selErr == nil && sel != "" {
			fmt.Printf(" [fzf] %s\n", sel)
			return sel, nil
		}
		return PromptLine(prompt)
	}
	return input, nil
}

// PromptLineWithFzf kept for backward compatibility; it delegates to
// PromptLineOrFzf (which reads the whole line and treats "/" as fzf trigger).
func PromptLineWithFzf(prompt string) (string, error) {
	return PromptLineOrFzf(prompt)
}

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
