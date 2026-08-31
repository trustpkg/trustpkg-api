package npm

func PackageWatchlist() []string {
	return append([]string(nil), packagesWatchlist[:]...)
}
