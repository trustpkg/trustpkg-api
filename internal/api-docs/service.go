package apiDocs

type Service struct {
	openAPIJSON []byte
	handlers    *handlers
}

func NewService(appURL string) (*Service, error) {
	configureSwaggerHost(appURL)

	openAPIJSON, err := generateOpenAPIJSON()
	if err != nil {
		return nil, err
	}

	service := &Service{openAPIJSON: openAPIJSON}
	service.handlers = newHandlers(service)

	return service, nil
}
