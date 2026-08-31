#!/usr/bin/env bash

set -euo pipefail

if [[ -z "${POSTMAN_API_KEY:-}" ]]; then
  echo "postman-sync: POSTMAN_API_KEY not set, skipping"
  exit 0
fi

if [[ -z "${POSTMAN_COLLECTION_UID:-}" ]]; then
  echo "postman-sync: POSTMAN_COLLECTION_UID not set, skipping"
  exit 0
fi

if [[ ! "${POSTMAN_COLLECTION_UID}" =~ ^[a-zA-Z0-9_-]+-[a-zA-Z0-9_-]+$ ]]; then
  echo "postman-sync: invalid POSTMAN_COLLECTION_UID format"
  exit 1
fi

collection_file="${POSTMAN_COLLECTION_FILE:-postman/trust-pkg-api.json}"
if [[ ! -f "$collection_file" ]]; then
  echo "postman-sync: collection file not found: $collection_file"
  exit 1
fi

tmp_file=$(mktemp --suffix=.json)
trap 'rm -f "$tmp_file"' EXIT

printf '{"collection":' > "$tmp_file"
cat "$collection_file" >> "$tmp_file"
printf '}' >> "$tmp_file"

response=$(curl --silent --show-error --write-out $'\n%{http_code}' \
  -X PUT \
  "https://api.getpostman.com/collections/${POSTMAN_COLLECTION_UID}" \
  -H "x-api-key: ${POSTMAN_API_KEY}" \
  -H "Content-Type: application/json" \
  --data-binary "@${tmp_file}")

http_code="${response##*$'\n'}"
body="${response%$'\n'*}"
if [[ "$http_code" != "200" ]]; then
  echo "postman-sync: failed (HTTP $http_code): $body"
  exit 1
fi

echo "postman-sync: done"