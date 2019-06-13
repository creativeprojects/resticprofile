#!/bin/sh

BASE_DIR=$(cd $(dirname "$0") && pwd)
cd ${BASE_DIR}

python3 src/resticprofile.py --config examples/profiles.conf $@
