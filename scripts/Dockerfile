FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend

WORKDIR /frontend-build/web

COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && corepack prepare pnpm@10.33.2 --activate && pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm release

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS backend

ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

RUN sed -i "s#https\?://dl-cdn.alpinelinux.org/alpine#${ALPINE_MIRROR}#g" /etc/apk/repositories || true
WORKDIR /backend-build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

ENV GOPROXY=${GOPROXY} \
    GOSUMDB=${GOSUMDB}

# Copy go mod files and download dependencies (cached layer)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code (use .dockerignore to exclude unnecessary files)
COPY . .

# Inject frontend build output generated in the dedicated frontend stage.
COPY --from=frontend /frontend-build/server/router/frontend/dist ./server/router/frontend/dist

ARG TARGETOS TARGETARCH VERSION COMMIT
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build \
      -trimpath \
      -ldflags="-s -w -extldflags '-static'" \
      -tags netgo,osusergo \
      -o memos \
      ./cmd/memos

# Use minimal Alpine with security updates
FROM alpine:3.21 AS monolithic

ARG ALPINE_MIRROR=https://mirrors.aliyun.com/alpine

RUN sed -i "s#https\?://dl-cdn.alpinelinux.org/alpine#${ALPINE_MIRROR}#g" /etc/apk/repositories || true

# Install runtime dependencies and create non-root user in single layer
RUN apk add --no-cache tzdata ca-certificates && \
    addgroup -g 10001 -S nonroot && \
    adduser -u 10001 -S -G nonroot -h /var/opt/memos nonroot && \
    mkdir -p /var/opt/memos /usr/local/memos && \
    chown -R nonroot:nonroot /var/opt/memos

# Copy binary and entrypoint to /usr/local/memos
COPY --from=backend /backend-build/memos /usr/local/memos/memos
COPY --from=backend --chmod=755 /backend-build/scripts/entrypoint.sh /usr/local/memos/entrypoint.sh

# Switch to non-root user
USER nonroot:nonroot

# Set working directory to the writable volume
WORKDIR /var/opt/memos

# Data directory
VOLUME /var/opt/memos

ENV TZ="UTC" \
    MEMOS_PORT="5230"

EXPOSE 5230

ENTRYPOINT ["/usr/local/memos/entrypoint.sh", "/usr/local/memos/memos"]
