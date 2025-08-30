#!/bin/bash
# Usage: ./inline_test.sh /path/to/image.png
f="$1"
enc=$(base64 -i "$f")
printf "\x1b]1337;File=inline=1;size=%d:%s\a\n" "$(wc -c <"$f")" "$enc"
