#!/bin/bash

set -e

source ./scripts/helpers/runLog.sh "Starting createTestingNpmDb script"

curl -u admin:password \
  -X PUT \
  http://localhost:3200/npm

source ./scripts/helpers/doneLog.sh "createTestingNpmDb script completed successfully"