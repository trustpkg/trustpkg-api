#!/bin/bash

set -e 

source scripts/helpers/runLog.sh "Running pre-commit checks..."

go vet ./... || {
    source scripts/helpers/errorLog.sh "Go vet failed. Please fix the errors before committing."
    exit 1
}

govulncheck -show verbose ./... || {
            source scripts/helpers/errorLog.sh "Govulncheck failed. Please fix the errors before committing."
            exit 1
}

source scripts/helpers/successLog.sh "Pre-commit checks passed."