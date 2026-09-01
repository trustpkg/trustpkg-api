package apiDocs

import (
	"encoding/json"
	"fmt"

	"github.com/trustpkg/trustpkg-api/docs"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

const passwordComplexityPattern = `^(?=.*[0-9])(?=.*[A-Z])(?=.*[^A-Za-z0-9]).{8,64}$`
const npmPackagePattern = `^\s*@?[a-z0-9][a-z0-9._-]*(\s+[a-z0-9][a-z0-9._-]*)?(/[a-z0-9][a-z0-9._-]*)?\s*$`

func applyPasswordPatternConstraints(openAPIDoc *openapi3.T) {
	if openAPIDoc == nil {
		return
	}

	for _, schemaRef := range openAPIDoc.Components.Schemas {
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}

		passwordRef, hasPassword := schemaRef.Value.Properties["password"]
		if !hasPassword || passwordRef == nil || passwordRef.Value == nil {
			continue
		}

		passwordRef.Value.Pattern = passwordComplexityPattern
	}
}

func generateOpenAPIJSON() ([]byte, error) {
	var swaggerDoc openapi2.T
	if err := json.Unmarshal([]byte(docs.SwaggerInfo.ReadDoc()), &swaggerDoc); err != nil {
		return nil, fmt.Errorf("unmarshal swagger document: %w", err)
	}

	openAPIDoc, err := openapi2conv.ToV3(&swaggerDoc)
	if err != nil {
		return nil, fmt.Errorf("convert swagger document to openapi: %w", err)
	}

	applyPasswordPatternConstraints(openAPIDoc)

	openAPIJSON, err := json.MarshalIndent(openAPIDoc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal openapi document: %w", err)
	}

	var document map[string]any
	if err := json.Unmarshal(openAPIJSON, &document); err != nil {
		return nil, fmt.Errorf("unmarshal generated openapi document: %w", err)
	}
	applyPackageSearchConstraints(document)

	openAPIJSON, err = json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal constrained openapi document: %w", err)
	}

	return openAPIJSON, nil
}

func applyPackageSearchConstraints(document map[string]any) {
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		return
	}
	searchPath, ok := paths["/packages/search"].(map[string]any)
	if !ok {
		return
	}
	searchOperation, ok := searchPath["get"].(map[string]any)
	if !ok {
		return
	}
	parameters, ok := searchOperation["parameters"].([]any)
	if ok {
		for _, rawParameter := range parameters {
			parameter, ok := rawParameter.(map[string]any)
			if !ok || parameter["name"] != "query" {
				continue
			}
			schema, ok := parameter["schema"].(map[string]any)
			if ok {
				schema["pattern"] = npmPackagePattern
			}
		}
	}

	components, ok := document["components"].(map[string]any)
	if !ok {
		return
	}
	schemas, ok := components["schemas"].(map[string]any)
	if !ok {
		return
	}
	searchDocument, ok := schemas["packagesApi.SearchDocument"].(map[string]any)
	if !ok {
		return
	}
	properties, ok := searchDocument["properties"].(map[string]any)
	if !ok {
		return
	}
	name, ok := properties["name"].(map[string]any)
	if ok {
		name["pattern"] = npmPackagePattern
	}
	vulnerabilities, ok := properties["vulnerabilities"].(map[string]any)
	if !ok {
		return
	}
	additionalProperties, ok := vulnerabilities["additionalProperties"].(map[string]any)
	if ok {
		additionalProperties["minimum"] = 1
	}
}
