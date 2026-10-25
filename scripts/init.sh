set -e

source scripts/helpers/runLog.sh "Initializing pre-push and pre-commit hook"

cp ./scripts/prePush.sh .git/hooks/pre-push
chmod +x .git/hooks/pre-push

cp ./scripts/preCommit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

source scripts/helpers/doneLog.sh "Initialization git hooks complete successfully"