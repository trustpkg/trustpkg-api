package npm

import "context"

func Pipeline() {
	ctx := context.Background()

	fetchNpmPackagesByCurrentSequence(ctx)
}
