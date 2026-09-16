# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
ARG VERSION=dev
RUN go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/secureenv-api ./cmd/secureenv-api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/secureenv-api /secureenv-api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/secureenv-api"]
