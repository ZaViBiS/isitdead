# --- build ---
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-s -w" -o /app/server ./cmd/server

# --- runtime ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 1000 app

WORKDIR /app
COPY --from=builder /app/server .

USER app
EXPOSE 8080

CMD ["./server"]
