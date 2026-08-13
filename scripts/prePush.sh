#!/bin/bash

set -e

source ./scripts/helpers/runLog.sh "Starting pre-push script"

source ./scripts/bumpVersion.sh

source ./scripts/helpers/doneLog.sh "Pre-push script completed successfully"