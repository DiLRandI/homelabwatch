FROM node:24-alpine@sha256:a0b9bf06e4e6193cf7a0f58816cc935ff8c2a908f81e6f1a95432d679c54fbfd AS web-build
WORKDIR /src
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json web/package.json
RUN corepack enable && pnpm install --frozen-lockfile
COPY web/ web/
RUN pnpm --dir web build

FROM golang:1.26.4-alpine@sha256:3ad57304ad93bbec8548a0437ad9e06a455660655d9af011d58b993f6f615648 AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=web-build /src/web/dist /src/web/dist
RUN CGO_ENABLED=0 go build -o /out/homelabwatch ./cmd/homelabwatch

FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce
RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 homelabwatch \
    && adduser -S -D -H -u 10001 -G homelabwatch homelabwatch \
    && mkdir -p /data \
    && chown homelabwatch:homelabwatch /data
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
USER 10001:10001
ENTRYPOINT ["/app/homelabwatch"]
