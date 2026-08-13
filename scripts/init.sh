set -e

source scripts/helpers/runLog.sh "Initializing pre-push hook"

cp ./scripts/prePush.sh .git/hooks/pre-push
chmod +x .git/hooks/pre-push

source scripts/helpers/doneLog.sh "Initialization git hooks complete successfully"