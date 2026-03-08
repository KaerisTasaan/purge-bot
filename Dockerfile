FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /purgebot ./cmd/purgebot

FROM alpine:3.21

RUN apk add --no-cache ca-certificates && mkdir -p /data && chown nobody:nobody /data

COPY --from=builder /purgebot /purgebot

USER nobody

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD ["/purgebot", "healthcheck"]

ENTRYPOINT ["/purgebot", "-db", "/data/database.db"]
