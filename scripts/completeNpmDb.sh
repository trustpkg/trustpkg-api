#!/usr/bin/env bash

set -u

COUCH_URL="http://localhost:3200"
COUCH_DB="npm"
COUCH_USER="admin"
COUCH_PASSWORD="password"

NPM_REGISTRY="https://registry.npmjs.org"

MAX_PARALLEL=5
MAX_RETRIES=3

source ./scripts/helpers/runLog.sh "Starting completeNpmDb script"

PACKAGES=(
    express
    lodash
    react
    axios
    typescript
    fastify
    next
    vite
    eslint
    prettier
)

TMP_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "$TMP_DIR"
}

trap cleanup EXIT


fetch_package() {
    local package="$1"
    local output="$TMP_DIR/$package.json"

    echo "Downloading: $package"

    local attempt=1

    while [ "$attempt" -le "$MAX_RETRIES" ]; do

        if curl \
            --fail \
            --silent \
            --show-error \
            --max-time 30 \
            "$NPM_REGISTRY/$package" \
            -o "$output"; then

            break
        fi

        echo "WARNING: failed to download $package (attempt $attempt/$MAX_RETRIES)" >&2

        rm -f "$output"

        if [ "$attempt" -eq "$MAX_RETRIES" ]; then
            echo "ERROR: skipping $package" >&2
            return 1
        fi

        sleep 2

        attempt=$((attempt + 1))
    done


    if ! jq empty "$output" >/dev/null 2>&1; then
        echo "ERROR: invalid JSON for $package" >&2
        rm -f "$output"
        return 1
    fi


    jq \
        --arg id "$package" \
        '
        ._id = $id |
        .source = "npm"
        ' \
        "$output" > "$output.tmp"

    mv "$output.tmp" "$output"

    echo "OK: $package"
}


upload_package() {
    local package="$1"
    local file="$TMP_DIR/$package.json"

    if [ ! -f "$file" ]; then
        return 0
    fi

    echo "Uploading: $package"

    local response

    response=$(curl \
        --fail \
        --silent \
        --show-error \
        -u "$COUCH_USER:$COUCH_PASSWORD" \
        -H "Content-Type: application/json" \
        -X PUT \
        "$COUCH_URL/$COUCH_DB/$package" \
        --data-binary "@$file" \
        2>&1)

    if [ $? -ne 0 ]; then
        echo "ERROR: failed to upload $package" >&2
        echo "$response" >&2
        return 1
    fi

    echo "SAVED: $package"
}


export TMP_DIR
export NPM_REGISTRY
export MAX_RETRIES

export -f fetch_package


echo "Downloading packages..."

printf '%s\n' "${PACKAGES[@]}" |
    xargs -n1 -P"$MAX_PARALLEL" \
    bash -c 'fetch_package "$1"' _


echo
echo "Downloading completed."
echo


echo "Uploading packages..."

SUCCESS=0
FAILED=0

for package in "${PACKAGES[@]}"; do

    if upload_package "$package"; then
        SUCCESS=$((SUCCESS + 1))
    else
        FAILED=$((FAILED + 1))
    fi

done


source ./scripts/helpers/doneLog.sh "completeNpmDb script completed successfully"