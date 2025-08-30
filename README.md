# termagick

A small interactive terminal image editor built in Go that uses ImageMagick (via the `gopkg.in/gographics/imagick.v3` binding) to apply image processing operations from your terminal. termagick exposes a curated set of ImageMagick operations with metadata-driven prompts so you can apply effects, transform images, and save results — all without leaving the terminal.

---

## Table of Contents

- Project Title
- Description
- Features
- Installation Instructions
  - macOS (Homebrew)
  - Debian / Ubuntu
  - Fedora / RHEL
  - Build the Go binary
- Usage
  - Running termagick
  - Interactive keys
  - Example session
- Configuration
  - Command metadata (JSON) and MetaStore
  - Environment / linker notes
- Dependencies
- Troubleshooting
- Contributing
- License

---

## Description

termagick is an interactive, terminal-based image editor. It reads an image into an ImageMagick `MagickWand` and allows you to apply a variety of transformations and filters (blur, sharpen, resize, rotate, posterize, composite, etc.) using interactive prompts.

It ships with in-code metadata describing each command's parameters and validation rules so the CLI can prompt for typed parameters and provide helpful hints. It integrates with `fzf` to provide a fuzzy command selector.

---

## Key Features

- Interactive terminal workflow for editing images.
- Metadata-driven command prompts with types, hints and examples.
- Many built-in operations mapped to ImageMagick functions (blur, resize, sharpen, sepia, posterize, composite, and more).
- `fzf`-backed command selector for fast, fuzzy command lookup.
- Save edited images to arbitrary output files.
- Inline terminal image preview support for compatible terminals (kitty graphics protocol, iTerm2 OSC 1337 inline images, and Sixel-capable terminals). The CLI attempts to display a PNG preview after commands are applied and will fallback between supported protocols when available.

---

## Installation Instructions

Prerequisites: a working Go toolchain and ImageMagick with development headers/libraries installed.

Note: this project uses the `gopkg.in/gographics/imagick.v3` binding which requires linking to the native ImageMagick libraries. Follow the instructions for your platform to install the native dependencies.

### macOS (Homebrew)

1. Install ImageMagick and pkg-config:
   - `brew install imagemagick pkg-config`
2. Install `fzf` for fuzzy selection:
   - `brew install fzf`
3. Build the binary (see "Build the Go binary" below).

### Debian / Ubuntu

1. Install OS packages:
   - `sudo apt-get update`
   - `sudo apt-get install -y imagemagick libmagickwand-dev pkg-config build-essential`
2. Install `fzf`:
   - `sudo apt-get install -y fzf` (or install from the project's recommended method)
3. Build the binary.

### Fedora / RHEL

1. Install development packages:
   - `sudo dnf install -y ImageMagick ImageMagick-devel pkgconfig`
2. Install `fzf`:
   - `sudo dnf install -y fzf`
3. Build the binary.

### Build the Go binary

From the repository root (the directory containing `go.mod`):

- To build:
  - `go build ./...`
  - This produces an executable (for example `termagick` if the module main package is in the repository root).
- Or to install to your GOPATH/bin:
  - `go install ./...`

Note: See `go.mod` which contains the module and required Go version information.

---

## Usage

Basic usage:

- Run the binary with an input image:
  - `./termagick path/to/input.jpg`

On startup the program loads the input image into memory and presents an interactive prompt.

If the running terminal supports inline image rendering, termagick will attempt to show an inline PNG preview of the current in-memory image after applying commands. Previewing prefers the kitty graphics protocol (kitty/kitty-compatible terminals), then iTerm2's OSC 1337 inline-file sequence, and finally falls back to Sixel rendering on terminals that support it. Preview output is optional and non-fatal — failures to render a preview are ignored by the main interactive flow.

Interactive keys:

- `/` — open the command selector.
- `s` — save the current in-memory image to a file (you will be prompted for a filename).
- `q` — quit the program.
- Other keys — ignored in the current interactive loop.

How command invocation works:

1. Press `/`.
2. Select a command (for example `blur`, `resize`, `posterize`).
3. The CLI will present parameter prompts based on metadata. Prompts include the parameter name, type, and hints/examples when available.
4. After entering parameters the command is validated and applied to the in-memory image.
5. Repeat commands as desired, and finally press `s` to save the edited image.

Example session (conceptual):

- Run:
  - `./termagick input.jpg`
- In the prompt:
  - Press `/`
  - Select `blur`
  - Enter `radius (px): 0.0`
  - Enter `sigma: 1.5`
  - The program prints `Applied blur`
  - Press `/`
  - Select `resize`
  - Enter `width: 1024`
  - Enter `height: 768`
  - Program prints `Applied resize`
  - Press `s`, enter `output.jpg`
  - Program prints `Saved to output.jpg`
  - Press `q` to exit

Command list and per-command help: the project contains an in-code `commands` slice (see `commands.go`) which lists each supported command, its description and parameter metadata. This metadata is used by the UI to prompt and validate.

---

## Configuration

Metadata and programmatic configuration:

- Built-in metadata:
  - The executable uses the in-code `commands` variable (in `commands.go`) and constructs a `MetaStore` via `NewMetaStore(commands)`.
- Loading metadata from JSON:
  - The package exposes `LoadCommandMetaFromFile` and `NewMetaStoreFromFile` (in `meta.go`) if you want to maintain command metadata externally in a JSON file.
  - The current CLI executable does not read metadata from disk by default; to use JSON metadata at runtime, modify `main.go` to call `NewMetaStoreFromFile(path)` and handle errors appropriately.
- fzf integration:
  - If `fzf` is installed and available in PATH, the code uses the function `SelectCommandWithFzf` (in `fzf.go`) to present commands to the user. Otherwise a simple text prompt fallback is used.
- Environment / linker notes:
  - When building and running, your system must be able to find and link the ImageMagick libraries. If you get linking or runtime errors, you may need to set:
    - `PKG_CONFIG_PATH` (so `pkg-config` finds ImageMagick .pc files during build)
    - `LD_LIBRARY_PATH` / `DYLD_LIBRARY_PATH` (if runtime cannot find shared libs) — platform-specific.

---

## Dependencies

The key dependencies for building and running termagick are:

- Go toolchain (see `go.mod` for the module and Go version declared in the project).
  - The project uses modules; run `go build` or `go install`.
- ImageMagick and its development headers (native library).
  - Header packages are platform-specific (`libmagickwand-dev`, `ImageMagick-devel`, etc.)
  - pkg-config is recommended to make the native binding build succeed.
- Go binding to ImageMagick:
  - `gopkg.in/gographics/imagick.v3` (declared in `go.mod`)
- Optional:
  - `fzf` — used for fuzzy command selection.

Files in the repository of note:

- `main.go` — main interactive loop and CLI behavior.
- `commands.go` — in-code command metadata (names, descriptions, params).
- `meta.go` — helper utilities for metadata, validation, and JSON loading.
- `imagemagick.go` — mapping from command names + args to ImageMagick `MagickWand` calls.
- `fzf.go` — optional `fzf` integration.

---

## Troubleshooting

- Build/link errors referencing ImageMagick symbols:
  - Ensure ImageMagick and its development package are installed (not just the CLI `convert`).
  - Ensure `pkg-config` can locate ImageMagick: try `pkg-config --modversion MagickWand` or similar. Adjust `PKG_CONFIG_PATH` if necessary.
  - You may need whitelist the -Xpreprocessor flag in your environment.
    ```
    export CGO_CFLAGS_ALLOW='-Xpreprocessor'
    ```
- Runtime errors about missing shared libraries:
  - On Linux, ensure the directory containing `libMagickWand.so` is in `LD_LIBRARY_PATH` or in the system linker configuration (`/etc/ld.so.conf.d/` + `ldconfig`).
  - On macOS, ensure `DYLD_LIBRARY_PATH` includes the appropriate path, or that Homebrew-installed libraries are properly linked.
- Command validation errors:
  - The CLI validates parameters (required/optional, numeric ranges, boolean parsing)… read the prompts carefully and follow examples shown in the prompt hints.
- fzf not invoked:
  - If you press `/` and see the fallback list prompt, ensure `fzf` is installed and in your `PATH`. The program calls `fzf` directly (see `fzf.go`).

- Preview not appearing or debugging preview issues:
  - Preview is best-effort and depends on your terminal supporting one of the protocols (kitty, iTerm2 OSC 1337, or Sixel). If no supported protocol is detected, no preview will be shown.
  - Enable preview debug logging to stderr with `PREVIEW_DEBUG=1`. Debug messages are prefixed with `termagick-preview:` and will show the detection and fallback decisions the previewer makes.
  - You can influence protocol detection and behavior with environment variables:
    - `SIXEL_PREVIEW=1` — force-enable Sixel capability detection if your terminal supports Sixel but heuristics do not detect it.
    - `KITTY_PREVIEW_COLS` and `KITTY_PREVIEW_ROWS` — optional sizing hints used by the kitty graphics placement logic to request a specific render area.
  - Detection uses common environment hints (for example `KITTY_WINDOW_ID`, `TERM`, `TERM_PROGRAM`, `ITERM_SESSION_ID`, `KONSOLE_VERSION`, and `WT_SESSION`) so ensure your terminal exposes standard variables. If preview fails despite appearing to be supported, enable `PREVIEW_DEBUG` to see why a protocol was skipped or errored.
  - Previews are non-blocking; failures to render will not interrupt the interactive editor, but the debug logs can help you diagnose and (if needed) force appropriate behavior via the env vars listed above.
