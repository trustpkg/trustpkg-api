package npm

type collectionState struct {
	areGathered         bool
	queue               []string
	lastPackage         *string
	allUnscopedGathered bool
}

type baseCollectionState struct {
	lastBasePackage *string
	scopedQueue     []string
	shouldRunScoped bool
}
