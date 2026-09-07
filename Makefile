include .env

OPENAPI_FILE ?= docs/openapi.json
OPENAPI_TYPES_FILE ?= docs/openapi.types.ts
OPENAPI_ZOD_FILE ?= docs/openapi.zod.ts
POSTMAN_COLLECTION_FILE ?= postman/trust-pkg-api.json
POSTMAN_BASE_URL ?= http://localhost:3001
POSTGRES_CONTAINER ?= trustpkg-api_db_1
BATCH_SIZE ?= 2000
PAUSE_SECONDS ?= 0.20

.PHONY: docs
docs:
	go run github.com/swaggo/swag/cmd/swag@v1.8.1 init -g cmd/api-service/main.go --parseInternal --output docs

.PHONY: openapi
openapi: docs
	npx -y swagger2openapi docs/swagger.json -o $(OPENAPI_FILE)
	node scripts/add-openapi-constraints.mjs $(OPENAPI_FILE)

.PHONY: api-types
api-types: openapi
	npx -y openapi-typescript $(OPENAPI_FILE) -o $(OPENAPI_TYPES_FILE)

.PHONY: api-zod
api-zod: openapi
	npx -y openapi-zod-client $(OPENAPI_FILE) -o $(OPENAPI_ZOD_FILE) --export-schemas --export-types --with-docs

.PHONY: postman-collection
postman-collection: openapi
	mkdir -p "$(dir $(POSTMAN_COLLECTION_FILE))"
	npx -y openapi-to-postmanv2 -s "$(OPENAPI_FILE)" -o "$(POSTMAN_COLLECTION_FILE)" -p
	node scripts/configure-postman-collection.mjs "$(POSTMAN_COLLECTION_FILE)" "$(POSTMAN_BASE_URL)"

.PHONY: postman-sync
postman-sync: postman-collection
	@POSTMAN_API_KEY="$(POSTMAN_API_KEY)" \
	POSTMAN_COLLECTION_UID="$(POSTMAN_COLLECTION_UID)" \
	POSTMAN_COLLECTION_FILE="$(POSTMAN_COLLECTION_FILE)" \
	bash scripts/postman-sync.sh

.PHONY: api-artifacts
api-artifacts: api-types api-zod postman-sync

.PHONY: postman-clean
postman-clean:
	rm -f "$(POSTMAN_COLLECTION_FILE)"

.PHONY: init
init:
	bash scripts/init.sh

.PHONY: build
build: docs
	bash scripts/build.sh

.PHONY: dropBin
dropBin:
	bash scripts/dropBin.sh

.PHONY: replicator-service
replicator-service:
	air -c configs/air/replicator-service.toml

.PHONY: api-service
api-service: docs
	air -c configs/air/api-service.toml

.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database $(DB_URL) up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database $(DB_URL) down

.PHONY: migrate-force
migrate-force:
	migrate -path migrations -database $(DB_URL) force $(version)

.PHONY: backfill-package-slugs
backfill-package-slugs: migrate-up
	podman exec -i "$(POSTGRES_CONTAINER)" psql -v batch_size="$(BATCH_SIZE)" -v pause_seconds="$(PAUSE_SECONDS)" -U "$${POSTGRES_USER:-user}" -d "$${POSTGRES_DB:-postgres}" -f - < scripts/backfill_packages_slug.sql

.PHONY: reindex-package-slugs
reindex-package-slugs:
	go run ./cmd/typesense-slug-reindex $(ARGS)

.PHONY: help
help:
	@echo "Available commands:"
	@echo "make docs                    - Generate Swagger documentation"
	@echo "make openapi                 - Generate OpenAPI 3 schema"
	@echo "make api-types               - Generate TypeScript types from OpenAPI"
	@echo "make api-zod                 - Generate Zod schemas and types from OpenAPI"
	@echo "make postman-collection      - Generate a Postman collection from OpenAPI"
	@echo "make postman-sync            - Generate and upload the Postman collection"
	@echo "make api-artifacts           - Generate OpenAPI, TypeScript, Zod, and Postman artifacts"
	@echo "make help                   	   - Show this help message"
	@echo "make init                   	   - Initialize pre-push hook"
	@echo "make build                  	   - Build the project"
	@echo "make dropBin                    - Drop the binary files"
	@echo "make replicator-service         - Run the replicator service with air in development mode" 
	@echo "make api-service                - Run api service"
	@echo "make migrate-up                 - Run database migrations up"
	@echo "make migrate-down               - Run database migrations down"
	@echo "make migrate-force              - Force database migration to a specific version (usage: make migrate-force version=<version_number>)"
	@echo "make backfill-package-slugs     - Backfill and uniquely index package slugs"
	@echo "                                (options: BATCH_SIZE=2000 PAUSE_SECONDS=0.20)"
	@echo "make reindex-package-slugs      - Populate slugs in the Typesense packages collection with parameters ARGS='-batch-size 128 -pause 250ms'"
	@echo "                                (options: ARGS='-batch-size 128 -pause 250ms')"