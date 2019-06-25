#!/bin/sh

BASE_DIR=$(cd $(dirname "$0") && pwd)
cd ${BASE_DIR}

PYTHONPATH=./src python3 -m pytest $@
