ARG GO_VERSION=1.26

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build
ARG SERVICE
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go run github.com/swaggo/swag/cmd/swag@v1.8.1 init -g cmd/api-service/main.go --parseInternal --output docs \
    && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/service /service
COPY --from=build /src/migrations /migrations
ENV MIGRATIONS_URL=file:///migrations
ENTRYPOINT ["/service"]
