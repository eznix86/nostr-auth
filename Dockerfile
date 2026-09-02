# syntax=docker/dockerfile:1.27

FROM oven/bun:1 AS web-deps
WORKDIR /app
COPY package.json bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache bun install --frozen-lockfile

FROM web-deps AS web-build
WORKDIR /app
COPY vite.config.js ./
COPY resources ./resources
RUN bun run build

FROM golang:1.26.1-alpine AS go-build
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY assets.go ./
COPY cmd ./cmd
COPY internal ./internal
COPY resources/images/images.json ./resources/images/images.json
COPY resources/views ./resources/views
COPY public ./public
COPY --from=web-build /app/public/build ./public/build
ARG TARGETOS=linux
ARG TARGETARCH
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags='-s -w' -o /out/nostr-auth ./cmd/web

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=go-build /out/nostr-auth /app/nostr-auth
EXPOSE 3000
ENTRYPOINT ["/app/nostr-auth"]
