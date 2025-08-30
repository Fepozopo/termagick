#!/bin/bash
# Usage: ./kitty_test.sh /path/to/image.png
f="$1"
if [ -z "$f" ]; then
  echo "usage: $0 file.png"
  exit 1
fi
# Check if the file exists
if [ ! -f "$f" ]; then
  echo "Error: File not found - $f"
  exit 1
fi
# Encode the image to base64 and send it to the Kitty terminal
b64=$(base64 -i "$f")
printf "\x1b_Ga=T,f=100,t=d;%s\x1b" "$b64"
