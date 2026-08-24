package adaptiveWorker

type JobController struct {
	AdjustReservationRatioHandler func(float64)
	AdjustSkippedCyclesHandler    func(int)
}
