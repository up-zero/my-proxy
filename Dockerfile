# syntax = docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o my-proxy-${TARGETOS}-${TARGETARCH}${TARGETVARIANT} .

FROM alpine:latest AS final-amd64
COPY --from=builder /app/my-proxy-linux-amd64 /usr/local/bin/my-proxy

FROM alpine:latest AS final-arm64
COPY --from=builder /app/my-proxy-linux-arm64 /usr/local/bin/my-proxy

FROM final-${TARGETARCH} AS final

RUN chmod +x /usr/local/bin/my-proxy

ENV GIN_MODE=release

EXPOSE 12312

CMD ["my-proxy", "serve"]
