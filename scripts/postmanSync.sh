#!/usr/bin/env bash

set -euo pipefail

if [[ -f ".env" ]]; then
  set -a
  source .env
  set +a
fi

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

COLLECTION_FILE="${POSTMAN_COLLECTION_FILE:-postman/morphyxis-mail-service.postman_collection.json}"

TMPFILE=$(mktemp --suffix=.json)
trap 'rm -f "$TMPFILE"' EXIT

if [[ ! -f "$COLLECTION_FILE" ]]; then
  echo "postman-sync: local file not found, pulling from Postman..."

  mkdir -p "$(dirname "$COLLECTION_FILE")"

  RESPONSE=$(curl -s -w "\n%{http_code}" \
    "https://api.getpostman.com/collections/${POSTMAN_COLLECTION_UID}" \
    -H "x-api-key: ${POSTMAN_API_KEY}")

  HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
  BODY=$(echo "$RESPONSE" | head -n-1)

  if [[ "$HTTP_CODE" != "200" ]]; then
    echo "postman-sync: pull failed (HTTP $HTTP_CODE): $BODY"
    exit 1
  fi

  echo "$BODY" | grep -o '"collection":{.*' | sed 's/"collection"://' | sed 's/}$//' > "$TMPFILE"

  if command -v jq &>/dev/null; then
    echo "$BODY" | jq '.collection' > "$COLLECTION_FILE"
  else
    echo "$BODY" > "$COLLECTION_FILE"
  fi

  echo "postman-sync: saved to $COLLECTION_FILE"
  exit 0
fi

echo "postman-sync: pushing to Postman (uid: ${POSTMAN_COLLECTION_UID})..."

FIRST_KEY=$(head -c 20 "$COLLECTION_FILE")
if [[ "$FIRST_KEY" == *'"collection"'* ]]; then
  cp "$COLLECTION_FILE" "$TMPFILE"
else
  printf '{"collection":' > "$TMPFILE"
  cat "$COLLECTION_FILE" >> "$TMPFILE"
  printf '}' >> "$TMPFILE"
fi

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X PUT \
  "https://api.getpostman.com/collections/${POSTMAN_COLLECTION_UID}" \
  -H "x-api-key: ${POSTMAN_API_KEY}" \
  -H "Content-Type: application/json" \
  --data-binary "@${TMPFILE}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [[ "$HTTP_CODE" != "200" ]]; then
  echo "postman-sync: push failed (HTTP $HTTP_CODE): $BODY"
  exit 1
fi

echo "postman-sync: done"
