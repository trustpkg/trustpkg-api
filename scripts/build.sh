#!/bin/bash

set -eu

if [ ! -d "bin" ]; then
  mkdir bin
fi

if [ ! -d "cmd" ]; then
  echo "cmd directory is missing"
  exit 1
else
  for dir in cmd/*/; do
    [ -d "$dir" ] || continue
    dirname=$(basename "$dir")
    mainFile="$dir/main.go"
    if [ -f "$mainFile" ]; then
      go build -o "bin/$dirname" "$mainFile"
      echo "binary $dirname created in bin directory"
    else
      echo "main file $mainFile not found, skipping"
    fi
  done
fi