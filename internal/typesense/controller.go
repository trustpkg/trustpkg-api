package typesense

import (
	"context"
	"fmt"

	typesenseApi "github.com/typesense/typesense-go/v4/typesense"
	"github.com/typesense/typesense-go/v4/typesense/api"
	"github.com/typesense/typesense-go/v4/typesense/api/pointer"
)

type Controller struct {
	ctx    context.Context
	client *typesenseApi.Client
}

func (controller *Controller) CreateCollection(collection *CollectionPayload) error {
	_, err := controller.client.Collections().Create(controller.ctx, collection.Collection)
	return err
}

func (controller *Controller) UpsertDocuments(collection *UpsertPayload) error {
	if collection == nil {
		return fmt.Errorf("Typesense upsert payload cannot be nil")
	}
	if len(collection.Documents) == 0 {
		return nil
	}

	action := api.Upsert
	params := &api.ImportDocumentsParams{
		Action: &action,
	}
	if collection.BatchSize > 0 {
		params.BatchSize = pointer.Int(collection.BatchSize)
	}

	_, err := controller.client.Collection(collection.Name).Documents().Import(controller.ctx, collection.Documents, params)
	if err != nil {
		return fmt.Errorf("upsert Typesense documents in %q: %w", collection.Name, err)
	}

	return nil
}

func (controller *Controller) SearchDocuments(collectionName string, params *api.SearchCollectionParams) (*api.SearchResult, error) {
	if collectionName == "" {
		return nil, fmt.Errorf("Typesense collection name cannot be empty")
	}
	if params == nil {
		return nil, fmt.Errorf("Typesense search params cannot be nil")
	}

	result, err := controller.client.Collection(collectionName).Documents().Search(controller.ctx, params)
	if err != nil {
		return nil, fmt.Errorf("search Typesense documents in %q: %w", collectionName, err)
	}

	return result, nil
}
