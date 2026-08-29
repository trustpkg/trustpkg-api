package osv

import "time"

type fetchPackageVulnerabilitiesPayload struct {
	name      string
	ecosystem *string
}

type SyncPackagePayload struct {
	Name      string
	Ecosystem *string
}

type OsvVuln struct {
	Affected         []OsvAffected       `json:"affected"`
	Aliases          []string            `json:"aliases"`
	DatabaseSpecific OsvDatabaseSpecific `json:"database_specific"`
	Details          string              `json:"details"`
	ID               string              `json:"id"`
	Modified         time.Time           `json:"modified"`
	Published        time.Time           `json:"published"`
	Withdrawn        *time.Time          `json:"withdrawn,omitempty"`
	References       []OsvReference      `json:"references"`
	SchemaVersion    string              `json:"schema_version"`
	Severity         []OsvSeverity       `json:"severity"`
	Summary          string              `json:"summary"`
}

type OsvAffected struct {
	DatabaseSpecific OsvAffectedDatabaseSpecific `json:"database_specific"`
	Package          OsvPackage                  `json:"package"`
	Ranges           []OsvRange                  `json:"ranges"`
}

type OsvAffectedDatabaseSpecific struct {
	Source string `json:"source"`
}

type OsvPackage struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
	PURL      string `json:"purl"`
}

type OsvRange struct {
	Events []OsvEvent `json:"events"`
	Type   string     `json:"type"`
}

type OsvEvent struct {
	Introduced   string `json:"introduced,omitempty"`
	Fixed        string `json:"fixed,omitempty"`
	LastAffected string `json:"last_affected,omitempty"`
}

type OsvDatabaseSpecific struct {
	CWEIDs           []string   `json:"cwe_ids"`
	GitHubReviewed   bool       `json:"github_reviewed"`
	GitHubReviewedAt *time.Time `json:"github_reviewed_at,omitempty"`
	NVDPublishedAt   *time.Time `json:"nvd_published_at,omitempty"`
	Severity         string     `json:"severity"`
}

type OsvReference struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type OsvSeverity struct {
	Score string `json:"score"`
	Type  string `json:"type"`
}

type fetchPackageVulnerabilitiesResponse struct {
	Vulns []OsvVuln `json:"vulns"`
}

type OsvQueryPackage struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

type OsvQuery struct {
	OsvPackage OsvQueryPackage `json:"package"`
}

type FetchOSVEcosystemDumpPayload struct {
	Ecosystem string
}

type OsvReplicationState struct {
	Ecosystem           string
	IsFirstSynchronised bool
	LastModifiedAt      *time.Time
	LastCheckedAt       *time.Time
	NextCheckAt         *time.Time
	LastSynchronizedAt  *time.Time
}

type ModifiedVulnerability struct {
	ID         string
	ModifiedAt time.Time
}
