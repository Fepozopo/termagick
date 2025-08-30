#!/bin/bash
# Usage: ./sixel_test.sh /path/to/image.png
if command -v img2sixel >/dev/null 2>&1; then
  img2sixel "$1"
else
  echo "img2sixel not found"
fi
