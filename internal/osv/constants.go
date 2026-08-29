package osv

import "time"

const (
	osvPackageVulnerabilitiesUrl = "https://api.osv.dev/v1/query"
	defaultOsvEcosystem          = "npm"

	osvDumpURL       = "https://storage.googleapis.com/osv-vulnerabilities"
	osvDumpURLSuffix = "all.zip"

	osvReplicationCooldown = 10 * time.Minute
	osvCursorOverlap       = time.Second
)
