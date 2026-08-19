package adaptiveWorker

type job struct {
	scheduledJob Job
	concurrency  int
	canRun       bool
}

func (job *job) inferShouldRunInCurrentCycle(currentCycle int) {
	nextCycle := currentCycle + 1
	cycleInterval := job.scheduledJob.SkippedCycles + 1

	job.canRun = nextCycle%cycleInterval == 0
}

func (job *job) calculateCycleConcurrency(
	availableConcurrency int,
	currentCycle int,
	runningJobs int,
) {
	job.inferShouldRunInCurrentCycle(currentCycle)

	if !job.canRun {
		job.concurrency = 0

		return
	}

	if availableConcurrency <= 0 || runningJobs <= 0 {
		job.concurrency = 0

		return
	}

	if availableConcurrency <= runningJobs {
		job.concurrency = 1

		return
	}

	remainingConcurrency := availableConcurrency - runningJobs

	job.concurrency = 1 + int(float64(remainingConcurrency)*job.scheduledJob.ReservationRatio)
}
