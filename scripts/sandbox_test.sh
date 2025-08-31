DIR=$(dirname "$(which fzf)")
PATH=$(echo "$PATH" | tr ':' '\n' | grep -v "^$DIR$" | paste -sd: -)
PATH=$PATH ./termagick
