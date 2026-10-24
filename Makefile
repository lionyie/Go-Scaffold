SHELL := /bin/bash

PNPM ?= pnpm

.PHONY: help init generate verify-generated verify-authz dev-server dev-web dev-admin build build-web build-admin build-plugins build-server copy-web test test-go test-web lint lint-web lint-proto vet db-reset module-add module-remove module-update module-list module-check-upstream example-add example-list example-remove rename docker clean

help:
	@echo "pnpm scripts are the source of truth; make is a thin alias."
	@echo ""
	@echo "  init / generate / verify-generated / verify-authz"
	@echo "  dev-server / dev-web / dev-admin"
	@echo "  build / build-web / build-admin / build-plugins / build-server"
	@echo "  test / lint / vet / db-reset"
	@echo "  module-add NAME=oauth-github  module-remove NAME=...  module-update NAME=..."
	@echo "  module-list  module-check-upstream"
	@echo "  example-add NAME=notes  example-list  example-remove NAME=notes"
	@echo "  rename NAME=myapp MODULE=github.com/me/myapp/services"
	@echo "  docker / clean"

init:
	$(PNPM) run init

generate:
	$(PNPM) run generate

verify-generated:
	$(PNPM) run verify-generated

verify-authz:
	$(PNPM) run verify-authz

dev-server:
	$(PNPM) run dev:server

dev-web:
	$(PNPM) run dev:web

dev-admin:
	$(PNPM) run dev:admin

build:
	$(PNPM) run build

build-web:
	$(PNPM) run build:web

build-admin:
	$(PNPM) run build:admin

build-plugins:
	$(PNPM) run build:plugins

build-server:
	$(PNPM) run build:server

copy-web:
	$(PNPM) exec node scripts/copy-web-dist.mjs

test: test-go test-web

test-go:
	cd services && go test ./...

test-web:
	$(PNPM) run test:web

lint: lint-proto lint-web vet

lint-proto:
	buf lint

lint-web:
	$(PNPM) run lint:web

vet:
	cd services && go vet ./...

db-reset:
	$(PNPM) run db:reset -- --yes

module-add:
	$(PNPM) run module:add -- $(NAME)

module-remove:
	$(PNPM) run module:remove -- $(NAME)

module-update:
	$(PNPM) run module:update -- $(NAME)

module-list:
	$(PNPM) run module:list

module-check-upstream:
	$(PNPM) run module:check-upstream

example-add:
	$(PNPM) run module:add -- example-$(NAME)

example-list:
	$(PNPM) run module:list

example-remove:
	$(PNPM) run module:remove -- example-$(NAME)

rename:
	$(PNPM) run rename -- --name $(NAME) $(if $(MODULE),--module $(MODULE),)

docker:
	docker build -f deploy/docker/Dockerfile -t scaffold:dev .

clean:
	$(PNPM) run clean
