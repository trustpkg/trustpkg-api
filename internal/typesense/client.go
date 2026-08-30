package typesense

import (
	"fmt"
	"os"

	typesenseApi "github.com/typesense/typesense-go/v4/typesense"
)

func newClient() *typesenseApi.Client {
	var serverSuffix string
	port := os.Getenv("TYPESENSE_PORT")
	if port != "" {
		serverSuffix = fmt.Sprintf(":%s", port)
	}

	client := typesenseApi.NewClient(
		typesenseApi.WithServer(os.Getenv("TYPESENSE_PROTOCOL")+"://"+os.Getenv("TYPESENSE_HOST")+serverSuffix),
		typesenseApi.WithAPIKey(os.Getenv("TYPESENSE_API_KEY")),
		typesenseApi.WithConnectionTimeout(connectionTimeout),
		typesenseApi.WithCircuitBreakerMaxRequests(maxSimultaneousAmountOfRequest),
		typesenseApi.WithCircuitBreakerInterval(circuitBreakerInterval),
		typesenseApi.WithCircuitBreakerTimeout(circuitBreakerTimeout),
	)

	return client
}
