package adaptiveWorker

func shouldRunInCycle(job Job, currentCycle, skippedCyclesOffset int) bool {
	nextCycle := currentCycle + 1
	cycleInterval := job.SkippedCycles + skippedCyclesOffset + 1

	return nextCycle%cycleInterval == 0
}
