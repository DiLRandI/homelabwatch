FROM node:24-alpine AS web-build
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=web-build /src/web/dist /src/web/dist
RUN CGO_ENABLED=0 go build -o /out/homelabwatch ./cmd/homelabwatch

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=go-build /out/homelabwatch /app/homelabwatch
COPY migrations /app/migrations
COPY --from=web-build /src/web/dist /app/web/dist
ENV LOG_LEVEL=info \
    HOMELABWATCH_LISTEN_ADDR=:8080 \
    HOMELABWATCH_DATA_DIR=/data \
    HOMELABWATCH_DB_PATH=/data/homelabwatch.db \
    HOMELABWATCH_STATIC_DIR=/app/web/dist
VOLUME ["/data"]
EXPOSE 8080
ENTRYPOINT ["/app/homelabwatch"]
