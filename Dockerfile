ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS build
ARG SERVICE
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go run github.com/swaggo/swag/cmd/swag@v1.8.1 init -g cmd/api-service/main.go --parseInternal --output docs \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/service /service
ENTRYPOINT ["/service"]
