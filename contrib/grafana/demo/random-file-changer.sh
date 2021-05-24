#!/usr/bin/env sh

target_path=/tmp/demo

[[ -d "$target_path" ]] \
  || mkdir "$target_path"

for i in 1 2 3 4 5 6 7 8 9 10; do
    file=$((1 + $RANDOM % 15))
    file="${target_path}/${file}.file"

    what=$(($RANDOM % 2))
    case $what in
        0) echo "Upd $file" ; echo $RANDOM > "$file" ;;
        1) echo "Del $file" ; [[ -e "$file" ]] && rm "$file" ;;
    esac
done

exit 0
