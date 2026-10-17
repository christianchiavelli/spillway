# syntax=docker/dockerfile:1

# Built with the Go of go.mod; the result is one static binary.
FROM golang:1.27.2-alpine3.24@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/api ./cmd/api

# Distroless: no shell, no package manager, a non-root user. The binary is its
# own health check, since there is no curl to run.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3
COPY --from=build /out/api /api
EXPOSE 8080 9090
HEALTHCHECK --interval=5s --timeout=4s --start-period=10s --retries=6 CMD ["/api", "healthcheck"]
ENTRYPOINT ["/api"]
CMD ["serve"]
