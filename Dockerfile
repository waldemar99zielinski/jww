# syntax=docker/dockerfile:1

FROM node:24-alpine AS web-build
WORKDIR /src/client/web
COPY client/web/package.json client/web/package-lock.json ./
RUN npm ci
COPY client/web/ ./
RUN npm run build

FROM golang:1.27-alpine AS server-build
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/jww-server ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S jww \
    && adduser -S -G jww jww
WORKDIR /app
COPY --from=server-build /out/jww-server /app/jww-server
COPY --from=web-build /src/client/web/dist /app/web

ENV HTTP_HOST=0.0.0.0 \
    HTTP_PORT=8000 \
    WEB_STATIC_DIR=/app/web
EXPOSE 8000
USER jww
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${HTTP_PORT}/api/v1/readiness" || exit 1
ENTRYPOINT ["/app/jww-server"]
