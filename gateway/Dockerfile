# ── Build stage ──────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /grafana-gateway \
    ./cmd/grafana-gateway

# ── Runtime stage ────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S gateway && \
    adduser -S -G gateway gateway

COPY --from=builder /grafana-gateway /usr/local/bin/grafana-gateway

USER gateway

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["grafana-gateway"]
