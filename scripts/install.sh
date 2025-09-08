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
  # try to extract library directory from pkg-config flags
  PKG_LIB_DIR=""
  # collect all -L paths reported by pkg-config and prefer the one that actually contains ImageMagick libs
  PKG_LIB_CANDIDATES="$(pkg-config --libs --static "$PKG" 2>/dev/null | sed -n 's/-L\([^ ]*\)/\1/p' || true)"
  for _d in $PKG_LIB_CANDIDATES; do
    if [ -d "$_d" ]; then
      if ls "$_d"/libMagickWand-7.Q16HDRI.so* >/dev/null 2>&1 || ls "$_d"/libMagickCore-7.Q16HDRI.so* >/dev/null 2>&1; then
        PKG_LIB_DIR="$_d"
        break
      fi
    fi
  done
  # fallback: if none of the candidates contained the expected libs, use the first candidate if present
  if [ -z "${PKG_LIB_DIR}" ] && [ -n "${PKG_LIB_CANDIDATES}" ]; then
    PKG_LIB_DIR="$(echo "$PKG_LIB_CANDIDATES" | awk '{print $1}')"
  fi
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
      PKG_LIB_DIR="$IM_DIR/lib"
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
    echo "The install script can try to register the ImageMagick lib directory with the system linker (ldconfig)."
    # If we detected a lib dir from pkg-config or Homebrew, offer to register it automatically
  if [ -n "${PKG_LIB_DIR:-}" ] && [ -d "$PKG_LIB_DIR" ] && (ls "$PKG_LIB_DIR"/libMagickWand-7.Q16HDRI.so* >/dev/null 2>&1 || ls "$PKG_LIB_DIR"/libMagickCore-7.Q16HDRI.so* >/dev/null 2>&1); then
      echo
      read -r -p "Register detected ImageMagick lib dir '$PKG_LIB_DIR' with ldconfig? [y/N]: " reg || true
      reg="${reg:-N}"
      if [[ "$reg" =~ ^[Yy]$ ]]; then
        CONF_FILE="/etc/ld.so.conf.d/termagick-imagemagick.conf"
        if [ "$(id -u)" -ne 0 ]; then
          echo "Adding $PKG_LIB_DIR to $CONF_FILE and running ldconfig (sudo will be used)"
          echo "$PKG_LIB_DIR" | sudo tee "$CONF_FILE" >/dev/null
          sudo ldconfig
        else
          echo "$PKG_LIB_DIR" | tee "$CONF_FILE" >/dev/null
          ldconfig
        fi
        echo "ldconfig updated. Re-checking library resolution..."
        UNRESOLVED=$(ldd "$BIN_PATH" 2>/dev/null | grep 'not found' || true)
        if [ -z "$UNRESOLVED" ]; then
          echo "All shared libraries resolved successfully."
        else
          echo "Still unresolved:"; echo "$UNRESOLVED"
        fi
      else
        echo "Detected lib dir '$PKG_LIB_DIR' does not appear to contain ImageMagick libs."
        echo "Skipping automatic ldconfig registration for that path."
        echo "You can register the correct library directory manually or provide it below."
        echo "Example: echo '/path/to/imagemagick/lib' | sudo tee /etc/ld.so.conf.d/termagick-imagemagick.conf && sudo ldconfig"
      fi
    else
      echo "Could not detect ImageMagick lib directory automatically."
      echo "If you know the lib directory, you can register it with ldconfig now."
      read -r -p "Enter ImageMagick lib directory to register (or leave empty to skip): " manual || true
      manual="${manual:-}"
      if [ -n "$manual" ] && [ -d "$manual" ]; then
        CONF_FILE="/etc/ld.so.conf.d/termagick-imagemagick.conf"
        if [ "$(id -u)" -ne 0 ]; then
          echo "$manual" | sudo tee "$CONF_FILE" >/dev/null
          sudo ldconfig
        else
          echo "$manual" | tee "$CONF_FILE" >/dev/null
          ldconfig
        fi
        echo "ldconfig updated."
      else
        echo "No directory provided or directory doesn't exist. Skipping registration."
        echo "You can set LD_LIBRARY_PATH or register the lib dir later as described in the README."
      fi
    fi
  else
    echo "All shared libraries resolved successfully."
  fi
else
  echo "Installed binary not found at $BIN_PATH; skipped ldd verification."
fi
