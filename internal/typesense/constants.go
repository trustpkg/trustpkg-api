package typesense

import "time"

const (
	connectionTimeout              = 5 * time.Second
	maxSimultaneousAmountOfRequest = 100
	circuitBreakerInterval         = 2 * time.Minute
	circuitBreakerTimeout          = 30 * time.Second

	PackagesCollection = "packages"
)
