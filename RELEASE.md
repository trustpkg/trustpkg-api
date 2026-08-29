# Release Notes

## OSV Vulnerability History

- Added immutable OSV vulnerability snapshots for every synchronized version.
- Added historical vulnerability-to-package relations for monthly reporting.
- Added withdrawn vulnerability tracking and idempotent snapshot persistence.
- Preserved existing npm popularity collection and API rate limits.
- Added PostgreSQL queries for monthly active vulnerability counts.

### Deployment note

Apply all database migrations before starting the replicator:

```bash
make migrate-up
make replicator-service
```
