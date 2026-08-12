#!/bin/bash

git fetch --tags

if git describe --tags --abbrev=0 >/dev/null 2>&1; then
  version=$(git describe --tags --abbrev=0)
  commit_range="$version..HEAD"
else
  version="0.0.0"
  echo "No tags found. Using default: $version"
  commit_range="HEAD"
fi

if ! git rev-parse --git-dir > /dev/null 2>&1; then
  echo "Error: Not a git repository or no commits found."
  exit 1
fi

function incrementVersionByConventionalCommits() {
  local current_version=$1
  local range=$2
  
  local commitMessages
  commitMessages=$(git log "$range" --pretty=format:%s 2>/dev/null)

  if [ -z "$commitMessages" ]; then
    echo "$current_version"
    return
  fi

  local majorCount
  majorCount=$(echo "$commitMessages" | grep -cE "^(feat|fix|perf|refactor|docs|style|test|chore)(\(.+\))?!: ")
  
  local minorCount
  minorCount=$(echo "$commitMessages" | grep -cE "^(feat)(\(.+\))?: ")
  
  local patchCount
  patchCount=$(echo "$commitMessages" | grep -cE "^(fix)(\(.+\))?: ")

  if [ "$majorCount" -gt 0 ]; then
    current_version=$(echo "$current_version" | awk -F. -v OFS=. '{$1++; $2=0; $3=0; print}')
  elif [ "$minorCount" -gt 0 ]; then
    current_version=$(echo "$current_version" | awk -F. -v OFS=. '{$2++; $3=0; print}')
  elif [ "$patchCount" -gt 0 ]; then
    current_version=$(echo "$current_version" | awk -F. -v OFS=. '{$3++; print}')
  fi

  echo "$current_version"
}

newVersion=$(incrementVersionByConventionalCommits "$version" "$commit_range")

if [ "$newVersion" != "$version" ]; then
  git tag "$newVersion"
  git push origin "$newVersion"
  
  echo "New tag created: $newVersion"
else
  echo "No version increment needed. Version remains: $version"
fi
