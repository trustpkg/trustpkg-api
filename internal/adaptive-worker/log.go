package adaptiveWorker

import "log"

const workerLogPrefix = "\033[36m[WORKER]\033[0m"

func workerLogf(format string, args ...any) {
	log.Printf(workerLogPrefix+" "+format, args...)
}
