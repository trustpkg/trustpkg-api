package databaseService

type MigrationStatus struct {
	CurrentVersion uint
	LatestVersion  uint
	HasCurrent     bool
	HasMigrations  bool
	Dirty          bool
}

func (status MigrationStatus) Pending() bool {
	return status.HasMigrations && (!status.HasCurrent || status.CurrentVersion < status.LatestVersion)
}

type Command struct {
	Action  string
	Version int
}
