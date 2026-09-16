# syntax=docker/dockerfile:1

# The build stage runs on the builder's native platform and cross-compiles for
# the target one, so multi-arch images build fast without emulation.
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd ./cmd
COPY internal ./internal
ARG TARGETOS TARGETARCH VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/secureenv-api ./cmd/secureenv-api

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
LABEL org.opencontainers.image.source="https://github.com/PoCInnovation/SecureEnv" \
      org.opencontainers.image.description="SecureEnv HTTP API in front of HashiCorp Vault" \
      org.opencontainers.image.licenses="Apache-2.0"
COPY --from=build /out/secureenv-api /secureenv-api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/secureenv-api"]
