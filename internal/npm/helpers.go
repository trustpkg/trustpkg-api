package npm

func getUniquePackages(packages []npmChange) []npmChange {
	unique := make(map[string]npmChange, len(packages))

	for _, item := range packages {
		unique[item.ID] = item
	}

	result := make([]npmChange, 0, len(unique))

	for _, item := range unique {
		result = append(result, item)
	}

	return result
}
