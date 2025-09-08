#!/usr/bin/env bash
set -euo pipefail

# Portable install helper for termagick
# Tries to use pkg-config for ImageMagick 7, falls back to Homebrew paths when available

PKG_V7="MagickWand-7.Q16HDRI"
PKG_LEGACY="MagickWand"

try_pkg() { pkg-config --exists "$1" 2>/dev/null; }

echo "Detecting ImageMagick via pkg-config..."
PKG=""
if try_pkg "$PKG_V7"; then
  PKG="$PKG_V7"
elif try_pkg "$PKG_LEGACY"; then
  PKG="$PKG_LEGACY"
else
  if command -v brew >/dev/null 2>&1; then
    BREW_PREFIX="$(brew --prefix)"
    export PKG_CONFIG_PATH="$BREW_PREFIX/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
    if try_pkg "$PKG_V7"; then
      PKG="$PKG_V7"
    elif try_pkg "$PKG_LEGACY"; then
      PKG="$PKG_LEGACY"
    fi
  fi
fi

if [ -n "$PKG" ]; then
  echo "Using pkg-config package: $PKG"
  export CGO_CFLAGS="$(pkg-config --cflags "$PKG")"
  export CGO_LDFLAGS="$(pkg-config --libs "$PKG")"
else
  echo "pkg-config couldn't find ImageMagick. Trying Homebrew Cellar heuristics..."
  if [ -n "${IM_PREFIX:-}" ]; then
    IM_DIR="${IM_PREFIX%/}"
  elif command -v brew >/dev/null 2>&1; then
    IM_PREFIX="$(brew --prefix)/Cellar/imagemagick"
    IM_DIR="$(ls -d "$IM_PREFIX"/* 2>/dev/null | tail -n1 || true)"
    if [ -n "$IM_DIR" ] && [ -d "$IM_DIR/lib" ]; then
      echo "Found Homebrew ImageMagick at: $IM_DIR"
      export CGO_CFLAGS="-I$IM_DIR/include/ImageMagick-7"
      export CGO_LDFLAGS="-L$IM_DIR/lib -lMagickWand-7.Q16HDRI -lMagickCore-7.Q16HDRI"
      export PKG_CONFIG_PATH="$IM_DIR/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
      # Offer to register library directory with the system linker (ldconfig)
      if [ "$(id -u)" -ne 0 ]; then
        echo
        read -r -p "Would you like to add $IM_DIR/lib to the system linker cache via sudo and ldconfig? [y/N]: " reply || true
        reply="${reply:-N}"
        if [[ "$reply" =~ ^[Yy]$ ]]; then
          echo "Adding $IM_DIR/lib to /etc/ld.so.conf.d/homebrew-imagemagick.conf and running ldconfig (sudo will be used)"
          echo "$IM_DIR/lib" | sudo tee /etc/ld.so.conf.d/homebrew-imagemagick.conf >/dev/null
          sudo ldconfig
          echo "ldconfig updated"
        else
          echo "Skipping ldconfig update. You can run the following as root to register the path later:" 
          echo "  echo '$IM_DIR/lib' | sudo tee /etc/ld.so.conf.d/homebrew-imagemagick.conf && sudo ldconfig"
        fi
      else
        # already root
        echo "$IM_DIR/lib" | tee /etc/ld.so.conf.d/homebrew-imagemagick.conf >/dev/null
        ldconfig
      fi
    else
      echo "Could not locate ImageMagick automatically. Install ImageMagick dev package or set CGO_CFLAGS/CGO_LDFLAGS manually."
      echo "Example: export CGO_CFLAGS='-I/path/to/include' ; export CGO_LDFLAGS='-L/path/to/lib -lMagickWand-7.Q16HDRI -lMagickCore-7.Q16HDRI'"
      exit 1
    fi
  else
    echo "No Homebrew detected and pkg-config failed. Please install ImageMagick dev packages or set CGO env vars manually."
    exit 1
  fi
fi

echo "Installing termagick via go..."
export CGO_CFLAGS_ALLOW='-Xpreprocessor'
# Install the CLI main package (module root doesn't contain a main package)
go install github.com/Fepozopo/termagick/cmd/termagick@latest
echo "Install complete. Ensure runtime linker can find ImageMagick libs (see README for LD_LIBRARY_PATH/ldconfig guidance)."

# CI-friendly verification: check the installed binary for unresolved shared libs
BIN_PATH="$(go env GOPATH 2>/dev/null || echo "$HOME/go")/bin/termagick"
if [ -x "$BIN_PATH" ]; then
  echo
  echo "Verifying runtime dependencies for: $BIN_PATH"
  UNRESOLVED=$(ldd "$BIN_PATH" 2>/dev/null | grep 'not found' || true)
  if [ -n "$UNRESOLVED" ]; then
    echo "WARNING: unresolved shared libraries detected:"
    echo "$UNRESOLVED"
    echo
    echo "If the unresolved libraries are ImageMagick libs, either set LD_LIBRARY_PATH to include the ImageMagick lib directory, or register the directory with the system linker via ldconfig (see README)."
  else
    echo "All shared libraries resolved successfully."
  fi
else
  echo "Installed binary not found at $BIN_PATH; skipped ldd verification."
fi
