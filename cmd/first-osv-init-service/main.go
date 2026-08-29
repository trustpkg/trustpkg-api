package main

import (
	"context"
	"log"
	"os"

	"github.com/trustpkg/trustpkg-api/db"
	"github.com/trustpkg/trustpkg-api/internal/osv"
)

const (
	ecosystem = "npm"
)

func main() {
	log.Println("Starting first osv initialization...")

	db.ConnectDb()
	ctx := context.Background()

	if err := osv.ReplicateEcosystem(ctx, osv.FetchOSVEcosystemDumpPayload{
		Ecosystem: ecosystem,
	}); err != nil {
		log.Printf("OSV replication failed: %v", err)

		os.Exit(1)
	} else {
		log.Println("OSV initialization process completed successfully.")
	}
}
