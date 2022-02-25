#!/usr/bin/env bash
SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd -P )"

TEMP_PATH="$(mktemp -d)"
if [[ -z ${TEMP_PATH} || ! -d "${TEMP_PATH}" ]] ; then
  exit 1
fi
export ROOT_PATH="${TEMP_PATH}"

trap 'cleanup remove ; cd "${SCRIPT_PATH}"' EXIT INT TERM

function cleanup() {
  echo "Cleaning '${TEMP_PATH}' $1"
  rm -rf "${TEMP_PATH}"
  [[ "$1" == "remove" ]] || mkdir -p "${TEMP_PATH}"
}

function invoke() {
  cd "${SCRIPT_PATH}" && ./post-install.sh
}

function setup() {
  cleanup ""
  local config="${TEMP_PATH}/etc/resticprofile"
  mkdir -p "${config}" \
  && cp -R *.conf *.rc conf.d profiles.d repositories.d templates "${config}" \
  && find "${config}" -name "*.conf" -exec mv {} {}.dist \;
  find "${config}"
}

# TODO tests
setup && invoke || exit 1
find "${TEMP_PATH}"

