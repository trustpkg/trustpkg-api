package npm

type npmRevision struct {
	Rev string `json:"rev"`
}

type npmChange struct {
	Seq     int64         `json:"seq"`
	ID      string        `json:"id"`
	Deleted bool          `json:"deleted"`
	Changes []npmRevision `json:"changes"`
}

type NpmChangesResponse struct {
	LastSeq int         `json:"last_seq"`
	Results []npmChange `json:"results"`
}

type dbInsertMarkDeprecatedBatchPayload struct {
	packages []string
}

type dbInsertPackagesBatchPayload struct {
	packages []string
}

type fetchNpmPackageDownloadsPayload struct {
	packages []string
	since    *string
}

type fetchNpmPackageDownloadsReturnType map[string]*FetchNpmPackageDownloadsResponse

type FetchNpmPackageDownloadsResponse struct {
	Downloads int64  `json:"downloads"`
	End       string `json:"end"`
	Package   string `json:"package"`
	Start     string `json:"start"`
}
