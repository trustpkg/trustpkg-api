include .env

.PHONY: init
init:
	bash scripts/init.sh

.PHONY: help
help:
	@echo "Available commands:"
	@echo "make help        - Show this help message"
	@echo "make init        - Initialize pre-push hook"