#!/bin/bash

set -eu

source ./scripts/helpers/runLog.sh "Starting build script"

if [ ! -d "bin" ]; then
  mkdir bin
fi

if [ ! -d "cmd" ]; then
  source ./scripts/helpers/errorLog.sh "cmd directory is missing"
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
      source ./scripts/helpers/errorLog.sh "main file $mainFile not found, skipping"
    fi
  done
fi

source ./scripts/helpers/doneLog.sh "Build script completed successfully"