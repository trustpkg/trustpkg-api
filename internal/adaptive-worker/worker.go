package adaptiveWorker

import (
	"context"

	"github.com/robfig/cron/v3"
)

type handler func()

func New(ctx context.Context) (controller *Controller, start handler) {
	newCron := cron.New(
		cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger),
		),
	)

	controller = &Controller{
		ctx:                 ctx,
		simultaneousWorkers: defaultMinWorkers,
		newCron:             newCron,
	}

	return controller, func() {
		controller.start()
	}
}
