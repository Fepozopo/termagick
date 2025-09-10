#!/usr/bin/env bash
set -euo pipefail

# Portable build helper for termagick
# Tries to use pkg-config for ImageMagick 7, falls back to Homebrew paths when available

export CGO_CFLAGS_ALLOW='-Xpreprocessor'

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

# Output path based on host
OUT_DIR="bin"
mkdir -p "$OUT_DIR"
OUT="$OUT_DIR/termagick-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"

echo "Building termagick -> $OUT"
go build -tags no_pkgconfig -o "$OUT" ./cmd/termagick
echo "Build complete: $OUT"
