package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptLine displays a prompt and reads a line of input from the user.
func PromptLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
