#!/bin/sh
set -e

MIXINS="
  database
  snapshot
"

PREFIX="${RESTICPROFILE_MIXINS_PREFIX:-"mixins"}"
TEMP_FILE="${TMPDIR:-/tmp}/.rp-mixin.tmp"
GIT_BRANCH="${RESTICPROFILE_BRANCH:-"master"}"
GIT_BASE_URL="${RESTICPROFILE_GIT_URL:-"https://raw.githubusercontent.com/creativeprojects/resticprofile/${GIT_BRANCH}/contrib/mixins"}"

move_download_to() {
  [ -s "$TEMP_FILE" ] && mv -f "$TEMP_FILE" "$1"
  return $?
}

download() {
  result=0
  if which -s curl ; then
    curl -fsL "$1" > "$TEMP_FILE" && move_download_to "$2"
    result=$?
  elif which -s wget ; then
    wget -nv -O "$TEMP_FILE" "$1" && move_download_to "$2"
    result=$?
  else
    echo "neither curl nor wget found, cannot load $1"
    result=1
  fi

  [ -e "$TEMP_FILE" ] && rm "$TEMP_FILE"
  return $result
}

download_all() {
  dir=""
  if [ -n "$1" ] && [ -d "$1" ] ; then
    dir="$1/"
    echo "downloading to $dir"
  fi
  for m in $MIXINS ; do
      url="$GIT_BASE_URL/${m}.yaml"
      dest="${dir}${PREFIX}-${m}.yaml"
      echo "getting $url > $dest"
      download "$url" "$dest" || echo "failed"
  done
}

download_all "$1"