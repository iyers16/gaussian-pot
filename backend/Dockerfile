FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run unit tests (game logic only — pure, no DB required).
RUN go test ./internal/game/...

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# ─── Runtime image ────────────────────────────────────────────────────────────
FROM alpine:3.20

# ca-certificates: needed for any outbound TLS; tzdata: correct time zones.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
