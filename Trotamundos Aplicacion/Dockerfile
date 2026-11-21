# syntax=docker/dockerfile:1

FROM golang:alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /trotamundos ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget && adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /trotamundos /app/trotamundos
COPY --from=builder /src/web /app/web

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/app/trotamundos"]
