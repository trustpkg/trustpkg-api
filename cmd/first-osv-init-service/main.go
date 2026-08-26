package main

import (
	"context"
	"log"

	"github.com/trustpkg/trustpkg-api/db"
	"github.com/trustpkg/trustpkg-api/internal/osv"
)

func main() {
	db.ConnectDb()

	if err := osv.ReplicateEcosystem(context.Background(), osv.FetchOSVEcosystemDumpPayload{
		Ecosystem: "npm",
	}); err != nil {
		log.Printf("OSV replication failed: %v", err)
	}
}
