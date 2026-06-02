FROM node:24-alpine AS web-build
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web ./
RUN npm run build

FROM golang:1.26-alpine AS go-build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-build /src/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/isitdead ./cmd/server

FROM alpine:3.22
WORKDIR /app
RUN addgroup -S isitdead && adduser -S isitdead -G isitdead
COPY --from=go-build /out/isitdead /app/isitdead
USER isitdead
EXPOSE 8080
ENTRYPOINT ["/app/isitdead"]
