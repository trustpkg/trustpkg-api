package npm

import "strings"

func isScopedPackage(packageName string) bool {
	return strings.HasPrefix(packageName, "@")
}

func getDividedPackagesByScope(packages []string) (scoped []string, unscoped []string) {
	for _, pkg := range packages {
		if isScopedPackage(pkg) {
			scoped = append(scoped, pkg)
		} else {
			unscoped = append(unscoped, pkg)
		}
	}

	return
}

func parseDataToPayload(packagesList fetchNpmPackageDownloadsReturnType) map[string]int64 {
	correct := make(map[string]int64, len(packagesList))
	for packageName, currentPackage := range packagesList {
		if currentPackage != nil {
			correct[packageName] = currentPackage.Downloads
		}
	}

	return correct
}
