#!/bin/sh
PAK_DIR="$(dirname "$0")"
cd "$PAK_DIR" || exit 1

export LD_LIBRARY_PATH=/usr/trimui/lib:$PAK_DIR/resources/lib

./aesthetics
