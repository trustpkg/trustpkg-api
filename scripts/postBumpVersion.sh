#!/usr/bin/env bash

set -e

function check_version_in_file() {
  local file="$1"
  local version="$2"

  if [[ ! -f "$file" ]]; then
    echo "File $file not found, skipping version check"
    return 0
  fi

  if ! grep -q "$version" "$file"; then
    echo "Version $version not found in $file"
    return 1
  fi
}

function set_version_in_file() {
  local file="$1"
  local version="$2"

  if [[ ! -f "$file" ]]; then
    echo "File $file not found, skipping version set"
    return 0
  fi

  sed -i -E "s/[0-9]+\.[0-9]+\.[0-9]+/$version/g" "$file"

  git add "$file"
  git commit -m "chore: update version in $file to $version" --no-verify
}

TAG=$(git describe --tags --abbrev=0) 
if [[ $TAG == "" ]]; then
  TAG="0.0.0"
fi

files=(
  "README.md"
  "./cmd/morphixis-mail-service/main.go"
)

for file in "${files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "$file not found, skipping"
    continue
  fi

  if check_version_in_file "$file" "$TAG"; then
    echo "Version $TAG found in $file, skipping"
  else
    echo "Version $TAG not found in $file, updating..."
    set_version_in_file "$file" "$TAG"
  fi
done

