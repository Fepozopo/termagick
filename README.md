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
- Updates & check-for-updates
- Configuration & Metadata
- Dependencies
- Troubleshooting
- Contributing
- License

---

## Project Overview

`termagick` is an interactive, terminal-based image editor. It loads an image into an ImageMagick `MagickWand` and allows you to apply transformations and filters (blur, sharpen, resize, rotate, posterize, composite, etc.) using interactive prompts driven by command metadata.

The CLI ships with built-in command metadata (see `commands.go`) and provides helpers to load metadata (see `meta.go`). Command prompts and validation are metadata-driven, so parameter types, required fields, enums and hints are shown to the user.

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
- Check-for-updates support triggered from the interactive UI (`u` key). See "Updates & check-for-updates" below for details.

---

## Dependencies
- Required:
  - `ImageMagick` command-line tools (for the `convert` binary, etc.) and libraries (version 7.X).
  - `pkg-config` (if you want to build from source).
  - A working Go toolchain.

- Optional but recommended:
  - `fzf` — used for fuzzy selection of commands and files.
  - Terminal with support for one of the following image protocols for inline previews:
    - `kitty` graphics protocol
    - `iTerm2` inline images (OSC 1337)
    - `sixel` (e.g. `mlterm`, `xterm` with sixel support, `mintty`, etc.)
  - Optional CLI tools for preview fallbacks (if your terminal does not support the above protocols):
    - `chafa`
    - `img2sixel`

If `fzf` is not installed, the program falls back to typed prompts. Similarly, if your terminal does not support inline image protocols, the program continues to function without previews.

---

## Installation

Once you have a working Go toolchain and ImageMagick installed, you can install `termagick` with `go install`.

Check if pkg-config is able to find the right ImageMagick include and libs:

```bash
pkg-config --cflags --libs MagickWand
```

Per the security update [https://groups.google.com/forum/#!topic/golang-announce/X7N1mvntnoU](https://groups.google.com/forum/#!topic/golang-announce/X7N1mvntnoU) you may need whitelist the -Xpreprocessor flag in your environment.

```bash
export CGO_CFLAGS_ALLOW='-Xpreprocessor'
```

Then install with:

```bash
go install github.com/Fepozopo/termagick@latest
```

You can also manually set the following environment variables so the Go linker can find the C libraries for ImageMagick.

```bash
export CGO_CFLAGS="$(pkg-config --cflags MagickWand-7.Q16HDRI)"
export CGO_LDFLAGS="$(pkg-config --libs MagickWand-7.Q16HDRI)"
go install -tags no_pkgconfig github.com/Fepozopo/termagick@latest
```

This will install it to your `$GOBIN` if set. Otherwise, it will install to `$GOPATH/bin` or `$HOME/go/bin`.

If you prefer, you can download prebuilt binaries for your platform from the [releases page](https://github.com/Fepozopo/termagick/releases).
Make sure to pick the right binary for your platform/architecture and place it in your `$PATH`.

---

## Scripts & runtime linking

This repo includes helper scripts in the `scripts/` directory to make building and installing more portable across systems that use Homebrew or system packages.

- `scripts/build.sh`: Detects ImageMagick via `pkg-config` (prefers `MagickWand-7.Q16HDRI`), falls back to Homebrew Cellar heuristics, sets `CGO_CFLAGS`/`CGO_LDFLAGS`, and builds a binary into `./bin/` named `termagick-<os>-<arch>`.
- `scripts/install.sh`: Similar detection logic; sets `CGO_*` env vars and runs `go install github.com/Fepozopo/termagick@latest`. When it detects a Homebrew Cellar path it will offer to register the library directory with the system linker (`ldconfig`) so the runtime loader can find `libMagickWand-7`.

Environment overrides and verification

- You can force a specific ImageMagick installation by setting `IM_PREFIX` to the ImageMagick installation lib directory (or its Cellar path). Example:
```bash
IM_PREFIX="/home/linuxbrew/.linuxbrew/Cellar/imagemagick/7.1.2-3" ./scripts/build.sh
IM_PREFIX="/home/linuxbrew/.linuxbrew/Cellar/imagemagick/7.1.2-3" ./scripts/install.sh
```

- `scripts/install.sh` performs a simple `ldd` check on the installed binary to report unresolved shared libraries and prints remediation steps.

Runtime linker notes (when you see "error while loading shared libraries: libMagickWand-7.Q16HDRI.so.10: cannot open shared object file: No such file or directory"):

- Quick test (no sudo):
```bash
export LD_LIBRARY_PATH="/path/to/im/lib:${LD_LIBRARY_PATH:-}"
./termagick
```

- Permanent system fix (requires sudo): add ImageMagick lib directory to the loader cache and run `ldconfig`:
```bash
echo "/path/to/im/lib" | sudo tee /etc/ld.so.conf.d/homebrew-imagemagick.conf
sudo ldconfig
```

Replace `/path/to/im/lib` with the directory containing `libMagickWand-7.Q16HDRI.so.*` (for Homebrew that is often `$(brew --prefix)/Cellar/imagemagick/<version>/lib`).

If you prefer not to modify system linker config, keep the `LD_LIBRARY_PATH` export in your shell profile.

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

Docker / Makefile build targets

This repository includes a `Makefile` with helper targets that use Docker BuildKit / buildx to build a Linux binary (and images) without requiring you to install Go or ImageMagick locally.

Prerequisites:

- Docker with BuildKit / buildx enabled (`docker buildx version`).
- If running on an arm64 host (Apple Silicon) and producing amd64 artifacts, buildx handles cross building.

Common targets / examples (run from repo root):

- Export a linux binary to `./out/termagick` using buildx local output:

```
make binary
# If you need a specific Go base image or arch:
make binary TARGET=linux/amd64 GOLANG_IMAGE=golang:1.25
```

- Fallback extraction (builds the `builder` stage for the requested platform, loads the image, then `docker cp` the binary out):

```
make extract-by-docker TARGET=linux/amd64 GOLANG_IMAGE=golang:1.25
# Result: ./out/termagick
```

- Build the runtime image (final stage of the Dockerfile):

```
make build
# optional override:
make build GOLANG_IMAGE=golang:1.25 TARGET=linux/amd64
```

- Build and push multi-arch image (requires docker login):

```
make multiarch IMAGE=youruser/termagick:tag PLATFORMS=linux/amd64,linux/arm64
```

Notes:

- If `make binary` fails due to host filesystem permissions when exporting `--output type=local`, try:
  - `mkdir -p out && chmod 0777 out` before running `make binary`, or
  - use `make extract-by-docker` which loads the built image and copies the binary out (more portable).
- You can override the Go base image used during the build by setting `GOLANG_IMAGE` on the make command line (for example `GOLANG_IMAGE=golang:1.24-bullseye`).

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
- `u` — check for updates (see "Updates & check-for-updates").
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

## Updates & check-for-updates

termagick includes a built-in update checker and an automatic updater helper. You can trigger an update check interactively by pressing the `u` key. The update logic:

- The running binary has a `Version` string (a `var Version = "..."` in `update.go`) that indicates the current version.
- When you trigger an update check, the program:
  - Queries the GitHub Releases API for releases of the repository `Fepozopo/termagick`.
  - Uses tolerant semver detection to find the highest valid semver release (it tolerates tags like `v1.2.3` or `1.2.3` and tries to extract semver substrings from the tag or release name).
  - Prefers published, non-prerelease, non-draft releases.
  - From the selected release, it attempts to pick a downloadable asset. The detector prefers assets whose names contain platform/arch hints (e.g. `darwin`, `linux`, `windows`, `amd64`, `arm64`) and otherwise falls back to the first asset.
- If a newer release with a downloadable asset is found:
  - The user is prompted to confirm the update (a simple `y/N` prompt).
  - On confirmation, the updater downloads the asset and tries to replace the running executable using a self-update helper.
  - After a successful update the updater attempts to restart the application by executing the new binary (via `syscall.Exec`). If that fails it will try to spawn the new binary as a child process, and if that also fails it informs the user to restart manually.
- If the release does not include a usable asset, the program informs you and instructs you to download manually from the releases page.
- If the `Version` string in the running binary is not parseable as semver, the updater will warn but still attempt to compare and detect newer releases using the release semver values.

Privacy & safety notes:

- The update check makes unauthenticated requests to the public GitHub Releases API. It does not upload usage data.
- The automatic update process replaces the running binary. If you prefer not to allow this, decline the update prompt or download releases manually from the GitHub releases page.

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

## Troubleshooting

- Build/link errors referencing ImageMagick symbols:
  - Ensure ImageMagick and its development package are installed (not just the `convert` CLI).
  - Ensure `pkg-config` can locate ImageMagick (`pkg-config --cflags --libs MagickWand`).
  - Ensure `CGO_CFLAGS` and `CGO_LDFLAGS` are set correctly if not using `pkg-config`.
  - Ensure `CGO_CFLAGS_ALLOW` includes `-Xpreprocessor` if needed.
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
- Update check / auto-update issues:
  - The update checker requires network access to `api.github.com` to query releases.
  - Automatic updates require a downloadable asset attached to the GitHub release. If the release has no suitable asset, the updater will instruct you to download manually.
  - If the automatic restart fails after updating, the updater will print instructions; you can restart manually.
  - Confirm the release assets include platform/arch hints in their filenames so the asset selection logic can pick a suitable binary.

---

## Contributing

Contributions are welcome. If you're adding commands, prefer adding metadata entries to `commands.go` (or provide a JSON metadata file and adjust the program to load it). When adding commands, include descriptive `Description`, `Params`, and validation metadata so the CLI can present clear prompts.

If you'd like to change startup behavior, note that `main.go` now prefers `SelectFileWithFzf` when invoked without an argument. Adjust or extend that behavior as needed.

If you modify the update logic or release artifact naming, update the documentation above so maintainers know how to publish releases that work with the in-app updater.

---

## License

This project is released under an open-source license. See the `LICENSE` file in the repository root for details.
