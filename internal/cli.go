package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func usage() {
	fmt.Println("Commands available:")
	fmt.Println("  /  - select and apply command")
	fmt.Println("  o  - open another image at runtime")
	fmt.Println("  s  - save current image")
	fmt.Println("  u  - check for updates")
	fmt.Println("  h  - show this help message")
	fmt.Println("  q  - quit")
}

func RunCLI() {
	var inputImagePath string
	if len(os.Args) >= 2 {
		inputImagePath = os.Args[1]
	} else {
		// Show usage information if no input image path is provided.
		inputImagePath = ""
	}

	// Use in-code commands metadata (compile-time)
	store := NewMetaStore(Commands)

	imagick.Initialize()
	defer imagick.Terminate()

	var wand *imagick.MagickWand
	// If an input path was provided, create a wand and read it. Otherwise leave wand nil.
	if inputImagePath != "" {
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
		if err := PreviewWand(wand); err == nil {
			if info, ierr := GetImageInfo(wand); ierr == nil {
				fmt.Println(info)
			}
		}
	} else {
		wand = nil
	}

	fmt.Println("Terminal Image Editor")
	usage()

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
			if wand == nil {
				fmt.Println("No image loaded. Press 'o' to open an image first, or provide an image path as the first argument.")
				continue
			}
			var commandName string
			name, err := SelectCommandWithFzf(Commands)
			if err != nil || name == "" {
				// fzf unavailable, returned nothing, or errored — fall back to a textual selection list.
				fmt.Println("Command selection (fallback):")
				for i, c := range Commands {
					fmt.Printf("  %d) %s - %s\n", i+1, c.Name, c.Description)
				}
				selection, _ := PromptLine("Enter number or command name (leave empty to cancel): ")
				if selection == "" {
					fmt.Println("selection cancelled")
					continue
				}
				// Try numeric selection first (1-based)
				if idx, perr := strconv.Atoi(selection); perr == nil {
					if idx < 1 || idx > len(Commands) {
						fmt.Println("invalid selection")
						continue
					}
					commandName = Commands[idx-1].Name
				} else {
					// Treat input as command name — perform case-insensitive exact or prefix matching.
					selLower := strings.ToLower(selection)
					found := ""
					// exact (case-insensitive) match
					for _, c := range Commands {
						if strings.ToLower(c.Name) == selLower {
							found = c.Name
							break
						}
					}
					// if not found, try prefix matches (case-insensitive)
					if found == "" {
						matches := []string{}
						for _, c := range Commands {
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
			for _, cmd := range Commands {
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
				metaCmd := GetCommandMetaByName(store.Commands, commandName)
				if metaCmd != nil {
					tooltip, _, _ := store.GetCommandHelp(commandName)
					fmt.Println("\n" + tooltip + "\n")
					rawArgs = make([]string, len(metaCmd.Params))
					for i, p := range metaCmd.Params {
						typeLabel := string(p.Type)
						if p.Type == ParamTypeEnum && len(p.EnumOptions) > 0 {
							typeLabel = fmt.Sprintf("enum(%s)", strings.Join(p.EnumOptions, "|"))
						}
						prompt := fmt.Sprintf("%s (%s): ", p.Name, typeLabel)

						var val string
						var perr error

						// If this parameter looks like a filesystem path or filename, prefer the interactive
						// PromptLineWithFzf which lets the user press '/' to invoke fzf or type normally.
						lowerName := strings.ToLower(p.Name)
						lowerHint := strings.ToLower(p.Hint)
						if p.Type == ParamTypeString && (strings.Contains(lowerName, "path") || strings.Contains(lowerName, "file") || strings.Contains(lowerHint, "path") || strings.Contains(lowerHint, "file")) {
							val, perr = PromptLineWithFzf(prompt)
							if perr != nil {
								fmt.Fprintf(os.Stderr, "input error: %v\n", perr)
								val = ""
							}
						} else {
							val, perr = PromptLine(prompt)
							if perr != nil {
								fmt.Fprintf(os.Stderr, "input error: %v\n", perr)
								val = ""
							}
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
					if err := PreviewWand(wand); err == nil {
						if info, ierr := GetImageInfo(wand); ierr == nil {
							fmt.Println(info)
						}
					}
					continue
				}
			}

			// Fallback legacy behavior: prompt using the simple CommandMeta.Params list and pass raw inputs.
			rawArgs = make([]string, len(selectedCmd.Params))
			for i, param := range selectedCmd.Params {
				prompt := fmt.Sprintf("Enter %s: ", param.Name)

				// Prefer PromptLineWithFzf for string params that look like file paths or filenames.
				var val string
				if param.Type == ParamTypeString {
					lowerName := strings.ToLower(param.Name)
					// No ParamMeta.Hint available here in legacy path, so only inspect name.
					if strings.Contains(lowerName, "path") || strings.Contains(lowerName, "file") {
						// Use the same buffered reader to support single-key '/' detection.
						v, perr := PromptLineWithFzfReader(reader, prompt)
						if perr != nil {
							fmt.Fprintf(os.Stderr, "input error: %v\n", perr)
							v = ""
						}
						val = v
						rawArgs[i] = val
						continue
					}
				}

				typed, _ := PromptLine(prompt)
				rawArgs[i] = typed
			}
			if err := ApplyCommand(wand, commandName, rawArgs); err != nil {
				fmt.Fprintf(os.Stderr, "apply command error: %v\n", err)
				continue
			}
			fmt.Printf("Applied %s\n", commandName)
			// Update inline terminal preview if available.
			if err := PreviewWand(wand); err == nil {
				if info, ierr := GetImageInfo(wand); ierr == nil {
					fmt.Println(info)
				}
			}

		case 's':
			out, _ := PromptLine("Enter output filename: ")
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
				newPath, _ = PromptLine("Enter path to image to open (leave empty to cancel): ")
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
			if err := PreviewWand(wand); err == nil {
				if info, ierr := GetImageInfo(wand); ierr == nil {
					fmt.Println(info)
				}
			}
			continue

		case 'u':
			// Trigger an update check (runs the goroutine in CheckForUpdates)
			err := CheckForUpdates()
			if err != nil {
				fmt.Fprintf(os.Stderr, "update check error: %v\n", err)
			}
			continue

		case 'h':
			usage()
			continue

		case 'q':
			fmt.Println("Exiting...")
			return

		default:
			// ignore other keys
		}
	}
}
