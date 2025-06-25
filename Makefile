.DEFAULT_GOAL := help
SHELL = /bin/sh

# LANGUAGE SPECIFICS

GO := go
GOARCH := $(shell $(GO) env GOARCH)
GOOS := $(shell $(GO) env GOOS)
GOPATH := $(shell $(GO) env GOPATH)

# GO_FILES := $(shell find . -name "*.go" | grep --invert-match "vendor" | grep --invert-match "_test.go")
# GO_PROJECT_NAME := server-services-next
# GO_PKG := gitlab.playpod.ir/alpha/backend/$(GO_PROJECT_NAME)
# GO_PKG_LIST := $(shell $(GO) list $(GO_PKG)/... | grep --invert-match "vendor")
GO_PKG_LIST := ./...
GO_TAGS := unit

MIGRATE_CI := false
MIGRATE_DB_ADDR := 127.0.0.1:26257
MIGRATE_DB_NAME := server
MIGRATE_DSN := "cockroachdb://root@$(MIGRATE_DB_ADDR)/$(MIGRATE_DB_NAME)?sslmode=disable"
MIGRATE_NAME := migrate_name
MIGRATE_VER := development
ifeq ($(MIGRATE_CI), true)
	MIGRATE_SOURCE := gitlab://$(GITLAB_USER):$(GITLAB_TOKEN)@$(GITLAB_URL)/$(PROJECT_ID)/db/$(MIGRATE_DB_NAME)/$(MIGRATE_VER)\#$(PROJECT_REF)
	MIGRATE_TAGS := cockroachdb,gitlab
else
	MIGRATE_SOURCE := file://db/$(MIGRATE_DB_NAME)/$(MIGRATE_VER)
	MIGRATE_TAGS := cockroachdb
endif

.PHONY: migrate-create
migrate-create: ## Migrate Create
	$(GO) install -tags=$(MIGRATE_TAGS) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GOPATH)/bin/migrate -database=$(MIGRATE_DSN) -source=$(MIGRATE_SOURCE) -verbose create -dir=db/$(MIGRATE_DB_NAME)/$(MIGRATE_VER) -ext=sql -seq $(MIGRATE_NAME)

.PHONY: migrate-down
migrate-down: ## Migrate Down
	$(GO) install -tags=$(MIGRATE_TAGS) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GOPATH)/bin/migrate -database=$(MIGRATE_DSN) -source=$(MIGRATE_SOURCE) -verbose down 1

.PHONY: migrate-drop
migrate-drop: ## Migrate Drop
	$(GO) install -tags=$(MIGRATE_TAGS) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GOPATH)/bin/migrate -database=$(MIGRATE_DSN) -source=$(MIGRATE_SOURCE) -verbose drop -f

.PHONY: migrate-up
migrate-up: ## Migrate Up
	$(GO) install -tags=$(MIGRATE_TAGS) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GOPATH)/bin/migrate -database=$(MIGRATE_DSN) -source=$(MIGRATE_SOURCE) -verbose up

.PHONY: migrate-version
migrate-version: ## Migrate Version
	$(GO) install -tags=$(MIGRATE_TAGS) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GOPATH)/bin/migrate -database=$(MIGRATE_DSN) -source=$(MIGRATE_SOURCE) -verbose version