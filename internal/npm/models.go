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

type dbDropPackagesBatchPayload struct {
	packages []string
}

type dbInsertPackagesBatchPayload struct {
	packages []string
}
