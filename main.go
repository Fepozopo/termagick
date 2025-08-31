package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func promptLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func usageAndExit(prog string) {
	fmt.Printf("Usage: %s <input-image>\n", prog)
	fmt.Println("Interactive terminal image editor:")
	fmt.Println("  /  - select and apply command")
	fmt.Println("  o  - open another image at runtime")
	fmt.Println("  s  - save current image")
	fmt.Println("  q  - quit")
	os.Exit(1)
}

func main() {
	var inputImagePath string
	if len(os.Args) >= 2 {
		inputImagePath = os.Args[1]
	} else {
		// Prefer fzf-based file selection when no argument is provided.
		// If fzf is unavailable, cancelled, or returns nothing, fall back to a typed prompt.
		selected, selErr := SelectFileWithFzf(".")
		if selErr != nil || selected == "" {
			p, _ := promptLine("Enter input image path (leave empty to exit): ")
			if p == "" {
				usageAndExit(os.Args[0])
			}
			inputImagePath = p
		} else {
			inputImagePath = selected
		}
	}

	// Use in-code commands metadata (compile-time)
	store := NewMetaStore(commands)

	imagick.Initialize()
	defer imagick.Terminate()

	var wand *imagick.MagickWand
	wand = imagick.NewMagickWand()
	// Defer a cleanup function that will destroy whatever wand is current at program exit.
	defer func() {
		if wand != nil {
			wand.Destroy()
		}
	}()
	if err := wand.ReadImage(inputImagePath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read image %s: %v\n", inputImagePath, err)
		os.Exit(1)
	}

	// Try to show an initial preview in compatible terminals.
	// Ignore errors here so preview remains optional.
	_ = PreviewWand(wand)

	fmt.Println("Terminal Image Editor")
	fmt.Println("Commands available, press '/' to select one, 'o' to open a different image, 's' to save, 'q' to quit")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		r, _, err := reader.ReadRune()
		if err != nil {
			fmt.Fprintf(os.Stderr, "read input error: %v\n", err)
			continue
		}

		switch r {
		case '/':
			var commandName string
			name, err := SelectCommandWithFzf(commands)
			if err != nil || name == "" {
				// fzf unavailable, returned nothing, or errored — fall back to a textual selection list.
				fmt.Println("Command selection (fallback):")
				for i, c := range commands {
					fmt.Printf("  %d) %s - %s\n", i+1, c.Name, c.Description)
				}
				selection, _ := promptLine("Enter number or command name (leave empty to cancel): ")
				if selection == "" {
					fmt.Println("selection cancelled")
					continue
				}
				// Try numeric selection first (1-based)
				if idx, perr := strconv.Atoi(selection); perr == nil {
					if idx < 1 || idx > len(commands) {
						fmt.Println("invalid selection")
						continue
					}
					commandName = commands[idx-1].Name
				} else {
					// Treat input as command name — perform case-insensitive exact or prefix matching.
					selLower := strings.ToLower(selection)
					found := ""
					// exact (case-insensitive) match
					for _, c := range commands {
						if strings.ToLower(c.Name) == selLower {
							found = c.Name
							break
						}
					}
					// if not found, try prefix matches (case-insensitive)
					if found == "" {
						matches := []string{}
						for _, c := range commands {
							if strings.HasPrefix(strings.ToLower(c.Name), selLower) {
								matches = append(matches, c.Name)
							}
						}
						if len(matches) == 1 {
							found = matches[0]
						} else if len(matches) > 1 {
							fmt.Println("ambiguous selection, candidates:")
							for _, m := range matches {
								fmt.Println("  " + m)
							}
							continue
						}
					}
					if found == "" {
						fmt.Printf("unknown command: %s\n", selection)
						continue
					}
					commandName = found
				}
			} else {
				commandName = name
			}

			// Find the CommandMeta definition (from commands.go)
			var selectedCmd CommandMeta
			for _, cmd := range commands {
				if strings.EqualFold(cmd.Name, commandName) {
					selectedCmd = cmd
					break
				}
			}
			if selectedCmd.Name == "" {
				fmt.Printf("unknown command: %s\n", commandName)
				continue
			}

			// If we have metadata for this command, use it to present helpful prompts,
			// otherwise fall back to simple prompts.
			var rawArgs []string
			if store != nil {
				// Attempt to find metadata entry
				metaCmd, metaFound := store.byName[commandName]
				if metaFound {
					tooltip, _, _ := store.GetCommandHelp(commandName)
					fmt.Println("\n" + tooltip + "\n")
					rawArgs = make([]string, len(metaCmd.Params))
					for i, p := range metaCmd.Params {
						typeLabel := string(p.Type)
						if p.Type == ParamTypeEnum && len(p.EnumOptions) > 0 {
							typeLabel = fmt.Sprintf("enum(%s)", strings.Join(p.EnumOptions, "|"))
						}
						prompt := fmt.Sprintf("%s (%s): ", p.Name, typeLabel)

						val, err := promptLine(prompt)
						if err != nil {
							fmt.Fprintf(os.Stderr, "input error: %v\n", err)
							val = ""
						}
						rawArgs[i] = val
					}

					// Normalize & validate args using the metadata-driven helper.
					normArgs, err := NormalizeArgs(store, commandName, rawArgs)
					if err != nil {
						fmt.Fprintf(os.Stderr, "input validation error: %v\n", err)
						fmt.Println("aborting command due to input errors")
						continue
					}

					// Apply command with normalized args
					if err := ApplyCommand(wand, commandName, normArgs); err != nil {
						fmt.Fprintf(os.Stderr, "apply command error: %v\n", err)
						continue
					}
					fmt.Printf("Applied %s\n", commandName)
					// Update inline terminal preview if available.
					_ = PreviewWand(wand)
					continue
				}
			}

			// Fallback legacy behavior: prompt using the simple CommandMeta.Params list and pass raw inputs.
			rawArgs = make([]string, len(selectedCmd.Params))
			for i, param := range selectedCmd.Params {
				prompt := fmt.Sprintf("Enter %s: ", param.Name)
				val, _ := promptLine(prompt)
				rawArgs[i] = val
			}
			if err := ApplyCommand(wand, commandName, rawArgs); err != nil {
				fmt.Fprintf(os.Stderr, "apply command error: %v\n", err)
				continue
			}
			fmt.Printf("Applied %s\n", commandName)
			// Update inline terminal preview if available.
			_ = PreviewWand(wand)

		case 's':
			out, _ := promptLine("Enter output filename: ")
			if out == "" {
				fmt.Println("no filename provided")
				continue
			}
			if err := wand.WriteImage(out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to write image: %v\n", err)
				continue
			}
			fmt.Printf("Saved to %s\n", out)

		case 'o':
			// Open another image at runtime. Prefer fzf-based file selection; fall back to typed path.
			selected, selErr := SelectFileWithFzf(".")
			var newPath string
			if selErr != nil || selected == "" {
				// fzf failed, was cancelled, or returned nothing — fall back to a typed path prompt.
				newPath, _ = promptLine("Enter path to image to open (leave empty to cancel): ")
				if newPath == "" {
					fmt.Println("open cancelled")
					continue
				}
			} else {
				newPath = selected
			}

			newWand := imagick.NewMagickWand()
			if err := newWand.ReadImage(newPath); err != nil {
				fmt.Fprintf(os.Stderr, "failed to read image %s: %v\n", newPath, err)
				newWand.Destroy()
				continue
			}
			// Destroy the current wand (if any) and replace it with the newly opened one.
			if wand != nil {
				wand.Destroy()
			}
			wand = newWand
			fmt.Printf("Opened %s\n", newPath)
			// Update inline terminal preview if available.
			_ = PreviewWand(wand)
			continue

		case 'q':
			fmt.Println("Exiting...")
			return

		default:
			// ignore other keys
		}
	}
}
