package typesense

import (
	typesenseApi "github.com/typesense/typesense-go/v4/typesense/api"
)

type CollectionPayload struct {
	Name       string                         `json:"name"`
	Collection *typesenseApi.CollectionSchema `json:"collection"`
}

type UpsertPayload struct {
	Name      string
	BatchSize int
	Documents []any
}
