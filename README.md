# trustpkg-api

API for tracking vulnerability occurrence frequencies.

## How to start

1. Initialize Git hooks

```bash
make init
```
2. Install Go

This step depends on your operating system.

For Fedora:

```bash
sudo dnf install golang
```

Verify the installation:

```bash
go version
```
Configure Go tools

If you use Bash, add the following to ~/.bashrc:

```bash
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
export PATH="$PATH:$GOBIN"
```

Then reload your shell configuration:

```bash
source ~/.bashrc
```

Install Air for automatic application reloading during development:

```bash
go install github.com/air-verse/air@latest
```

3. Run the service

```bash
make replicator-service
```


## Replicator service

The replicator service is the background process started from `cmd/replicator-service`. At startup it
opens connections to PostgreSQL and Redis, then starts the adaptive worker. PostgreSQL is the main
database for packages, vulnerabilities, replication cursors, and collected package data. Redis stores
temporary and durable processing state, including worker cycles and queues of changes waiting to be
processed.

The service manages three background pipelines. The npm replication pipeline connects to the npm
database change feed, reads new package changes, and saves its current replication position. The npm
collector takes packages from the pending work queue, downloads the required package data from the npm
registry, and writes the results to PostgreSQL. Keeping replication and package collection as separate
jobs prevents a large upstream batch from blocking the rest of the service.

### Reindex package slugs in Typesense

After adding the `slug` field to an existing Typesense collection, run the standalone resumable
reindex command. It uses a separate checkpoint and does not change the normal Typesense pipeline
state:

```bash
make reindex-package-slugs ARGS='-batch-size 128 -pause 250ms'
```

The default checkpoint is `tmp/typesense-packages-slug-reindex.json`. If the command is interrupted,
run the same command again to continue. Delete the checkpoint, or pass `-reset`, only to start the
reindex from the beginning. The normal replicator may remain running because both operations use
idempotent upserts.

The OSV pipeline connects to the OSV data service for vulnerability feeds. It imports the complete
ecosystem dump when an ecosystem is initialized, then uses the incremental feed for later updates.
Changed vulnerability IDs are placed in Redis and processed one record per callback before being stored
in PostgreSQL. An item remains available when processing fails, so the next callback can retry it. A
cooldown also prevents the service from repeatedly downloading a complete OSV dump.

The adaptive worker checks scheduled jobs every second. Individual jobs can skip worker cycles when
they should run less often, and the worker can delay new work when CPU, system load, or memory usage
indicates pressure. Its cycle state is persisted in Redis, so scheduling can continue consistently
after a restart. The service logs use timestamps and module prefixes such as `[WORKER]`,
`[NPM COLLECTOR]`, `[NPM_REPLICATOR]`, and `[OSV Collector]` to make each pipeline easy to identify.

4. pick up Postgres db

- podman

```bash
podman compose up -d
```

 - or docker

```bash
docker compose up -d
```

## OSV replication

OSV data replication follows the [OSV data dump documentation](https://google.github.io/osv.dev/data/).
The first synchronization for an ecosystem downloads its complete dump:

```text
https://storage.googleapis.com/osv-vulnerabilities/<ecosystem>/all.zip
```

After the first successful synchronization, the service checks the ecosystem feed every 10 minutes:

```text
https://storage.googleapis.com/osv-vulnerabilities/<ecosystem>/modified_id.csv
```

Only IDs listed after the stored `last_modified_at` cursor are downloaded as JSON records. A one-second
cursor overlap prevents records with equal or near-identical timestamps from being skipped. Reprocessing
the overlap is safe because vulnerability writes are idempotent.

The synchronization state is stored in `osv_replication_state`, separately for every ecosystem. The first
full import sets `is_first_synchronised` only after all records have been committed. Incremental IDs are
stored in a Redis queue, and each worker callback processes one vulnerability at a time. This limits
external requests and PostgreSQL work while keeping failed items available for retry.

Every imported OSV version is also stored in `vulnerability_history`, together with its affected package
data in `vulnerability_history_affected`. These tables are append-only: changes to affected version
ranges and withdrawn records remain available for historical monthly reports. The existing
`vulnerabilities` and `vulnerability_affected` tables continue to contain the current read model.

Run `make migrate-up` before starting the service to create the history tables. The first history
migration records the current database state as a baseline; it cannot reconstruct versions that were
overwritten before the migration was installed. Subsequent OSV updates are recorded in the same
PostgreSQL transaction as the current-state update, so a failed update is retried from Redis without
leaving a partial history record. Generating a report reads PostgreSQL only and does not call OSV or npm.

OSV records with a `withdrawn` timestamp are retained and marked with `vulnerabilities.withdrawn_at`.
If an ID from `modified_id.csv` returns HTTP 404, an existing local record is marked as withdrawn as well.

Apply migrations before starting the service:

```bash
make migrate-up
```

## Package search API

The package search endpoint is available at `GET /api/packages/search`.

Query parameters:

- `query` (required): npm package name, 1-214 characters. Scoped names such as
	`@scope/package` are supported.
- `limit` (optional): number of results from 1 to 20. The default is 10.

Interactive Swagger documentation is available at `/api/swagger/`, and the OpenAPI 3
document is available at `/api/openapi.json`.

Generate the documentation and client artifacts with:

```bash
make api-artifacts
```

This generates `docs/swagger.json`, `docs/openapi.json`, `docs/openapi.types.ts`, and
`docs/openapi.zod.ts`. The OpenAPI schema is the source of truth for the generated
TypeScript types and Zod schemas, including required fields, enums, ranges, and the npm
package-name pattern.
The `docs/` directory is generated locally and ignored by Git.

The Postman collection is generated from the same OpenAPI document:

```bash
make postman-collection
```

To update an existing collection, set `POSTMAN_COLLECTION_UID` in `.env` and run:

```bash
make postman-sync
```

When `POSTMAN_COLLECTION_UID` is empty, synchronization is skipped without an error.
