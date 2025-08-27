package main

import (
	"bytes"
	"fmt"
	"os/exec"
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
