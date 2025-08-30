package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SelectCommandWithFzf displays a list of commands (using CommandMeta) in fzf and returns the selected command name.
func SelectCommandWithFzf(commands []CommandMeta) (string, error) {
	var b strings.Builder
	for _, c := range commands {
		// format as "name: description"
		b.WriteString(fmt.Sprintf("%s: %s\n", c.Name, c.Description))
	}

	cmd := exec.Command("fzf")
	cmd.Stdin = strings.NewReader(b.String())

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running fzf: %w", err)
	}

	selection := strings.TrimSpace(out.String())
	parts := strings.SplitN(selection, ":", 2)
	if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
		return strings.TrimSpace(parts[0]), nil
	}

	return "", fmt.Errorf("no command selected")
}

// SelectFileWithFzf launches fzf with a list of common image files found under startDir.
// It returns the full path of the selected file or an error if selection failed.
//
// Note: This implementation shells out to `find` piped into `fzf`. It requires both
// `find` and `fzf` to be available in PATH. startDir may be "." or any directory path.
func SelectFileWithFzf(startDir string) (string, error) {
	// Quote the directory to safely handle spaces/special chars.
	quotedDir := strconv.Quote(startDir)

	// Build a shell pipeline that finds image files and feeds them into fzf.
	// The percent sign in the format string is escaped as %%.
	cmdStr := fmt.Sprintf("find %s -type f \\( -iname '*.jpg' -o -iname '*.jpeg' -o -iname '*.png' -o -iname '*.gif' -o -iname '*.tif' -o -iname '*.tiff' \\) | fzf --height 40%% --border --prompt='Files> '", quotedDir)

	cmd := exec.Command("bash", "-lc", cmdStr)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running fzf for files: %w", err)
	}

	selection := strings.TrimSpace(out.String())
	if selection == "" {
		return "", fmt.Errorf("no file selected")
	}
	return selection, nil
}
