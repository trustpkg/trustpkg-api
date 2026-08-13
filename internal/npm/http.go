package npm

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

func fetchNpmPackagesByCurrentSequence(ctx context.Context) error {
	client := &http.Client{}

	sequence, err := selectLastSequence(ctx)
	if err != nil {
		return err
	}

	url := npmUrl + "_changes?since=" + strconv.Itoa(sequence) + "&limit=" + limit + "&include_docs=" + shouldIncludeDocs

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Accept-Encoding", "gzip")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	fmt.Println("response: ", response)

	if response.StatusCode == 200 {
		insertLastSequence(ctx, sequence)
	}

	return nil
}
