#!/bin/bash

set -euo pipefail

# ALPM calls this script from the root directory, with respect to the
# `--root` flag.
export HOLO_ROOT_DIR="."

comm -12 \
     <(# changed files come from stdin (without leading slash)
       sort | sed 's+^+/+') \
     <(# enumerate files that are managed by holo-files
       holo scan --short | sed -n 's/^file://p' | sort) \
| sed 's/^/file:/' | xargs -r holo apply
