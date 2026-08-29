package npm

const (
	npmReplicationUrl string = "https://replicate.npmjs.com/registry/_changes"
	limit             int    = 10000

	npmDownloadUrl           string = "https://api.npmjs.org/downloads/point"
	defaultNpmDownloadPeriod string = "last-month"
	npmDownloadBatchLimit    int    = 128
)
