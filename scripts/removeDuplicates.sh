#!/bin/bash

set -euo pipefail

file="$1"
tmp=$(mktemp)

awk '!seen[$0]++' "$file" > "$tmp"
mv "$tmp" "$file"

echo "Removed duplicates from $file"