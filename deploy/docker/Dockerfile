# syntax=docker/dockerfile:1.7

FROM node:24-alpine AS web
WORKDIR /src
RUN corepack enable
COPY . .
RUN pnpm install --frozen-lockfile
RUN pnpm --filter @scaffold/web build && pnpm --filter @scaffold/admin build && node scripts/build-plugins.mjs
RUN node scripts/copy-web-dist.mjs
RUN mkdir -p /out/plugins && \
    for d in /src/runtime-plugins/*/; do \
      [ -f "$d/plugin.json" ] || continue; \
      name=$(basename "$d"); \
      mkdir -p "/out/plugins/$name"; \
      cp "$d/plugin.json" "/out/plugins/$name/"; \
      [ -d "$d/dist" ] && cp -r "$d/dist" "/out/plugins/$name/"; \
      [ -d "$d/server" ] && cp -r "$d/server" "/out/plugins/$name/"; \
    done

FROM golang:1.27-alpine AS build
WORKDIR /src/services
COPY services/go.mod services/go.sum ./
RUN go mod download
COPY services ./
COPY --from=web /src/services/internal/webui/dist ./internal/webui/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X main.Version=${VERSION}" \
    -o /out/scaffold ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache git ca-certificates && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/scaffold /app/scaffold
COPY configs /app/configs
COPY --from=web /out/plugins /app/runtime-plugins
RUN mkdir -p /app/data/plugins && chown -R 10001:10001 /app/data
EXPOSE 8000 9000
USER 10001:10001
ENTRYPOINT ["/app/scaffold"]
CMD ["-conf", "/app/configs/docker.yaml"]
