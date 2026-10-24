# syntax=docker/dockerfile:1.7

FROM node:24-alpine AS web
WORKDIR /src
RUN corepack enable
COPY . .
RUN pnpm install --frozen-lockfile
RUN pnpm --filter @scaffold/web build && pnpm --filter @scaffold/admin build && node scripts/build-plugins.mjs
RUN node scripts/copy-web-dist.mjs

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

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/scaffold /app/scaffold
COPY configs /app/configs
COPY --from=web /src/runtime-plugins /app/runtime-plugins
EXPOSE 8000 9000
USER nonroot:nonroot
ENTRYPOINT ["/app/scaffold"]
CMD ["-conf", "/app/configs/docker.yaml"]
