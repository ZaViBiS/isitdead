FROM node:24-alpine AS web-builder

WORKDIR /app/web

COPY web/package*.json ./
RUN npm ci

COPY web ./
RUN npm run build

FROM golang:1.26.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
RUN go build -v -o ./bin/isitdead ./cmd/server

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/bin/isitdead ./isitdead

EXPOSE 8080
CMD ["./isitdead"]
