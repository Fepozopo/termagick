# termagick

A small interactive terminal image editor built in Go that uses ImageMagick (via the `gopkg.in/gographics/imagick.v3` binding) to apply image processing operations from your terminal. termagick exposes a curated set of ImageMagick operations with metadata-driven prompts so you can apply effects, transform images, and save results — all without leaving the terminal.

This repository contains the terminal UI and the command metadata used to present typed prompts and validation rules. The CLI integrates with `fzf` for fuzzy selection both for commands and (optionally) to pick files at startup or when opening a new image at runtime.

---

## Table of Contents

- Project Overview
- Key Features (recent changes highlighted)
- Installation
  - macOS (Homebrew)
  - Debian / Ubuntu
  - Fedora / RHEL
  - Build the Go binary
- Usage
  - Starting the program (fzf file selection)
  - Interactive keys (including `o` to open another image)
  - Command invocation and improved prompts
  - Example session
- Configuration & Metadata
- Dependencies
- Troubleshooting
- Contributing
- License

---

## Project Overview

`termagick` is an interactive, terminal-based image editor. It loads an image into an ImageMagick `MagickWand` and allows you to apply transformations and filters (blur, sharpen, resize, rotate, posterize, composite, etc.) using interactive prompts driven by command metadata.

The CLI ships with built-in command metadata (see `commands.go`) and provides helpers to load metadata from JSON (see `meta.go`). Command prompts and validation are metadata-driven, so parameter types, required fields, enums and hints are shown to the user.

---

## Key Features (recent)

- Interactive terminal workflow for editing images.
- At startup, when no input path is provided, termagick now attempts an `fzf`-backed file selection (`SelectFileWithFzf`). If `fzf` is not available or the selection is cancelled, the program falls back to a typed prompt.
- Open another image at runtime with the `o` key (also prefers `fzf`).
- Metadata-driven command prompts with types, hints and examples. Prompts are improved to show types (including enum options) and the metadata tooltip before prompting for parameters.
- `fzf`-backed command selector for fast, fuzzy command lookup (`SelectCommandWithFzf` in `fzf.go`). If `fzf` is not available, falls back to a typed prompt.
- Inline terminal image preview support for compatible terminals (kitty graphics protocol, iTerm2 OSC 1337 inline images, and Sixel-capable terminals). Previewing prefers kitty, then iTerm2, then Sixel. If none are supported, it attempts to use chafa.
- Save edited images to arbitrary output files.
- Preview is non-blocking and best-effort; failures do not interrupt the interactive flow.

---

## Installation

---

## Building from Source

Prerequisites: a working Go toolchain and ImageMagick with development headers/libraries installed.

> Note: `gopkg.in/gographics/imagick.v3` requires linking to the native ImageMagick libraries. Install native dependencies for your platform.

### macOS (Homebrew)

1. Install ImageMagick and pkg-config:
   - `brew install imagemagick pkg-config`
2. Install `fzf` (optional but recommended):
   - `brew install fzf`
3. Build:
   - See "Build the Go binary" below.

### Debian / Ubuntu

1. Install OS packages:
   - `sudo apt-get update`
   - `sudo apt-get install -y imagemagick libmagickwand-dev pkg-config build-essential`
2. Install `fzf`:
   - `sudo apt-get install -y fzf` (or follow the project's recommended install)
3. Build the binary.

### Fedora / RHEL

1. Install development packages:
   - `sudo dnf install -y ImageMagick ImageMagick-devel pkgconfig`
2. Install `fzf`:
   - `sudo dnf install -y fzf`
3. Build the binary.

### Build the Go binary

From the repository root (where `go.mod` is located):

- To build:
  - `go build ./...`
  - This produces executable(s) (for example `termagick` if the main package is in the repo root).
- To install to `GOBIN`:
  - `go install ./...`

See `go.mod` for module and Go version information.

---

## Usage

Basic invocation:

- Run the binary, optionally with an input image:
  - `./termagick path/to/input.jpg`
  - If you omit the input path, `termagick` prefers an `fzf`-backed file selection. If `fzf` is not present or the user cancels, you'll be prompted to type a path.

On startup the program loads the chosen image into memory and presents an interactive prompt. The current in-memory image is previewed (if the terminal supports a protocol) after commands are applied.

Interactive keys (in the interactive prompt):

- `/` — open the command selector (fzf-backed if available). Falls back to a typed prompt if `fzf` is not found.
- `o` — open another image at runtime (prefers `fzf` for selection; falls back to typed path).
- `s` — save the current in-memory image to a file (you will be prompted for a filename).
- `q` — quit the program.
- Other keys — ignored in the current interactive loop.

How command invocation works (improved prompts):

1. Press `/`.
2. Select a command (e.g. `blur`, `resize`, `posterize`) using `fzf` or the text fallback.
3. The CLI prints a generated tooltip describing the command (derived from `commands.go` metadata) and then prompts for parameters.
   - Each prompt shows the parameter name and a type label. For enums, the available options are shown (e.g. `enum(low|medium|high)`).
4. After entering parameters they are validated using the metadata rules; invalid inputs abort the command and return an informative error.
5. When validation passes, the command is applied to the in-memory `MagickWand`.
6. Preview is updated (best-effort).
7. Repeat as desired and press `s` to save.

Example session (conceptual):

- Run:
  - `./termagick` (no argument)
  - `fzf` appears to choose a file; you select `images/input.jpg`
- In the prompt:
  - Press `/`
  - Select `blur`
  - See tooltip and prompts:
    - `radius (float): 0.0`
    - `sigma (float): 1.5`
  - Program prints `Applied blur`
  - Press `/`
  - Select `resize`
  - Enter `width (int): 1024`
  - Enter `height (int): 768`
  - Program prints `Applied resize`
  - Press `s`, enter `output.jpg`
  - Program prints `Saved to output.jpg`
  - Press `q` to exit

---

## Configuration & Metadata

- Built-in metadata:
  - The executable uses the in-code `commands` variable (in `commands.go`) and constructs a `MetaStore` via `NewMetaStore(commands)` (see `meta.go`).
- Loading metadata from JSON:
  - Helpers `LoadCommandMetaFromFile` and `NewMetaStoreFromFile` are available in `meta.go` if you want to store command metadata externally.
  - By default the CLI uses the in-code metadata. To use external JSON metadata at runtime, modify `main.go` to call `NewMetaStoreFromFile(path)` and handle errors.
- fzf integration:
  - If `fzf` is installed and in `PATH`, `SelectCommandWithFzf` and `SelectFileWithFzf` will be used for command and file selection respectively. Otherwise a text prompt fallback is used.

Preview / terminal rendering notes:

- Previews are best-effort and optional. The previewer prefers the kitty graphics protocol, then iTerm2 OSC 1337 inline-file sequences, finally Sixel for compatible terminals.
- Control preview behavior with environment variables:
  - `PREVIEW_DEBUG=1` — enable debug logging from the previewer (helpful for diagnosing which protocol was chosen and why one failed).
  - `SIXEL_PREVIEW=1` — force-enable Sixel detection if your terminal supports Sixel but heuristics miss it.
  - `KITTY_PREVIEW_COLS` / `KITTY_PREVIEW_ROWS` — sizing hints for kitty placement logic.
- Preview-related logic is implemented in `terminal_preview.go`. Debug logging and detection follow environment heuristics and common terminal environment variables.

---

## Dependencies

### Run Dependencies

Note, there are no true runtime dependencies; if `fzf` is not installed, the program falls back to typed prompts. Similarly, if your terminal does not support inline image protocols, the program continues to function without previews.

- Optional but recommended:
  - `fzf` — used for fuzzy selection of commands and files.
  - Terminal with support for one of the following image protocols for inline previews:
    - `kitty` graphics protocol
    - `iTerm2` inline images (OSC 1337)
    - `sixel` (e.g. `mlterm`, `xterm` with sixel support, `mintty`, etc.)
  - Optional CLI tools for preview fallbacks (if your terminal does not support the above protocols):
    - `chafa`
    - `img2sixel`

### Build Dependencies

- Go toolchain (modules; see `go.mod`).
- ImageMagick native libraries and development headers.
- Go binding:
  - `gopkg.in/gographics/imagick.v3` (declared in `go.mod`).

Files of interest in this repo:

- `main.go` — main interactive loop, startup file-selection behavior, and key handling (including the `o` key to open another image).
- `commands.go` — built-in command metadata.
- `meta.go` — metadata helpers, `MetaStore`, JSON loading and validation helpers.
- `imagemagick.go` — mapping from command names + args to ImageMagick `MagickWand` calls.
- `fzf.go` — `fzf` integration for command and file selection.
- `terminal_preview.go` — inline preview helpers and protocol detection.

---

## Troubleshooting

- Build/link errors referencing ImageMagick symbols:
  - Ensure ImageMagick and its development package are installed (not just the `convert` CLI).
  - Ensure `pkg-config` can locate ImageMagick (`pkg-config --modversion MagickWand`).
  - If you get CGO-related issues, verify environment variables and that your system can find the headers and libraries.
- Runtime errors about missing shared libraries:
  - On Linux, make sure the directory containing `libMagickWand.so` is in `LD_LIBRARY_PATH` or installed in a standard location and registered via `ldconfig`.
  - On macOS, ensure `DYLD_LIBRARY_PATH` includes the right location or that Homebrew-installed libraries are properly linked.
- `fzf` not invoked:
  - If you press `/` or use startup file selection and the `fzf` UI does not appear, ensure `fzf` is installed and available in your `PATH`.
- Preview not appearing:
  - Previews depend on terminal protocol support and environment variables. Use `PREVIEW_DEBUG=1` to see diagnostic output from the previewer.
  - If your terminal supports Sixel but detection fails, set `SIXEL_PREVIEW=1` to force-enable it.
  - Kitty placement size can be influenced with `KITTY_PREVIEW_COLS` / `KITTY_PREVIEW_ROWS`.

---

## Contributing

Contributions are welcome. If you're adding commands, prefer adding metadata entries to `commands.go` (or provide a JSON metadata file and adjust the program to load it). When adding commands, include descriptive `Description`, `Params`, and validation metadata so the CLI can present clear prompts.

If you'd like to change startup behavior, note that `main.go` now prefers `SelectFileWithFzf` when invoked without an argument. Adjust or extend that behavior as needed.

---

## License

This project is released under an open-source license. See the `LICENSE` file in the repository root for details.
