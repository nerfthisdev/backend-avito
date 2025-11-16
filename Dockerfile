# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /out/pr-assigner ./cmd/service
RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.22.0

FROM debian:bookworm-slim AS runtime
WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates bash \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/pr-assigner /usr/local/bin/pr-assigner
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations /app/migrations
COPY docker/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
