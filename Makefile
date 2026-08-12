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

.PHONY: help
help:
	@echo "Available commands:"
	@echo "make help               - Show this help message"
	@echo "make init               - Initialize pre-push hook"
	@echo "make build              - Build the project"
	@echo "make dropBin     	   - Drop the binary files"
	@echo "make replicator-service - Run the replicator service with air in development mode" 