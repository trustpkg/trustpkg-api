package typesense

import (
	"context"
)

func NewInstance(ctx context.Context) *Controller {
	client := newClient()
	if isHealth, error := client.Health(ctx, connectionTimeout); error != nil {
		typesenseLog.Printf("Something gone wrong - Connection with Typesense Failed: %v", error)
	} else {
		if isHealth {
			typesenseLog.Print("Connection with Typesense established Successfully")
		} else {
			typesenseLog.Print("Connection with Typesense Failed")
		}
	}

	return &Controller{
		ctx:    ctx,
		client: client,
	}
}
