include .env

.PHONY: init
init:
	bash scripts/init.sh

.PHONY: build
build:
	bash scripts/build.sh

.PHONY: dropBin
dropBin:
	bash scripts/dropBin.sh

.PHONY: replicator-service
replicator-service:
	air -c configs/air/replicator-service.toml

.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database $(DB_URL) up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database $(DB_URL) down

.PHONY: migrate-force
migrate-force:
	migrate -path migrations -database $(DB_URL) force $(version)

.PHONY: complete-testing-npm-db
complete-testing-npm-db:
	bash scripts/completeNpmDb.sh

.PHONY: create-testing-npm-db
create-testing-npm-db:
	bash scripts/createTestingNpmDb.sh

.PHONY: help
help:
	@echo "Available commands:"
	@echo "make help                   	   - Show this help message"
	@echo "make init                   	   - Initialize pre-push hook"
	@echo "make build                  	   - Build the project"
	@echo "make dropBin                    - Drop the binary files"
	@echo "make replicator-service         - Run the replicator service with air in development mode" 
	@echo "make migrate-up                 - Run database migrations up"
	@echo "make migrate-down               - Run database migrations down"
	@echo "make migrate-force              - Force database migration to a specific version (usage: make migrate-force version=<version_number>)"
	@echo "make complete-testing-npm-db    - Complete testing npm database"
	@echo "make create-testing-npm-db      - Create testing npm database"