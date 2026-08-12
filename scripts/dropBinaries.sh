#!/bin/bash

set -e

shopt -s nullglob
filesInBinDir=(bin/*)
shopt -u nullglob

if [ ${#filesInBinDir[@]} -eq 0 ]; then 
  echo "there are no files in bin directory"
  exit 0
fi

rm -rf bin/* || echo "there are no files in bin directory"

for file in "${filesInBinDir[@]}"; do  
  filename=$(basename "$file")
  echo "file $filename removed from bin directory"
done

echo "all files removed from bin directory"