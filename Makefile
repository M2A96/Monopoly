.DEFAULT_GOAL := help
SHELL = /bin/sh

# GIT SPECIFICS

GIT_HOOKS = .git/hooks/commit-msg .git/hooks/pre-commit .git/hooks/pre-push .git/hooks/prepare-commit-msg

$(GIT_HOOKS): .git/hooks/%: .githooks/%

.githooks/%:
	touch $@

.git/hooks/%:
	cp $< $@

.PHONY: remove-git-configs
remove-git-configs: ## Remove Git Configs
	echo "remove-git-configs"

.PHONY: add-git-configs
add-git-configs: remove-git-configs ## Add Git Configs
	git config --global advice.skippedCherryPicks false
	git config --global branch.autosetuprebase always
	git config --global color.branch true
	git config --global color.diff true
	git config --global color.interactive true
	git config --global color.status true
	git config --global color.ui true
	git config --global commit.gpgsign true
	git config --global core.autocrlf input
	git config --global core.editor 'code --wait'
	git config --global diff.tool code
	git config --global difftool.code.cmd 'code --diff $$LOCAL $$REMOTE --wait'
	git config --global gpg.program gpg
	git config --global before.defaultbranch main
	git config --global log.date relative
	git config --global merge.tool code
	git config --global mergetool.code.cmd 'code --wait $$MERGED'
	git config --global pull.default current
	git config --global pull.rebase true
	git config --global push.autoSetupRemote true
	git config --global push.default current
	git config --global rebase.autostash true
	git config --global rerere.enabled true
	git config --global stash.showpatch true
	git config --global tag.gpgsign true

.PHONY: remove-git-hooks
remove-git-hooks: ## Remove Git Hooks
	rm -fr $(GIT_HOOKS)

.PHONY: add-git-hooks
add-git-hooks: remove-git-hooks $(GIT_HOOKS) ## Add Git Hooks

.PHONY: remove-git
remove-git: remove-git-configs remove-git-hooks ## Remove Git Configs & Hooks

.PHONY: add-git
add-git: add-git-configs add-git-hooks ## Add Git Configs & Hooks

.PHONY: help
help: ## Help
	@grep --extended-regexp "^[a-zA-Z_-]+:.*?## .*$$" $(MAKEFILE_LIST) \
| sort \
| awk 'BEGIN { FS = ":.*?## " }; { printf "\033[36m%-33s\033[0m %s\n", $$1, $$2 }'

# LANGUAGE SPECIFICS

GO := go
GOARCH := $(shell $(GO) env GOARCH)
GOOS := $(shell $(GO) env GOOS)
GOPATH := $(shell $(GO) env GOPATH)

# GO_FILES := $(shell find . -name "*.go" | grep --invert-match "vendor" | grep --invert-match "_test.go")
# GO_PROJECT_NAME := game-backend
# GO_PKG := gitlab.playpod.ir/alpha/backend/$(GO_PROJECT_NAME)
# GO_PKG_LIST := $(shell $(GO) list $(GO_PKG)/... | grep --invert-match "vendor")
GO_PKG_LIST := ./...
GO_TAGS := unit

BUF_CI := false
BUF_REF := $(shell git tag --list --sort=-creatordate | head -n 1)
ifeq ($(BUF_CI), true)
	BUF_BREAKING_AGAINST := $(CI_REPOSITORY_URL)\#branch=main,ref=$(BUF_REF),subdir=api
else
	BUF_BREAKING_AGAINST := .git\#branch=main,ref=$(BUF_REF),subdir=api
endif
K6_ADDRESS := 127.0.0.1:9090
K6_BASE_URL := http://127.0.0.1:7070
K6_PACKAGE := game_imdb_v1
K6_SLEEP_DURATION := 0.1
K6_TESTMAIL_API_KEY := 00ec60bc-a660-4490-a01f-41377b6f171e
K6_TESTMAIL_NAMESPACE := x7gw2
K6_USER_ID := 1e2d9f38-2777-4ee2-ac3b-b3a108f81a30
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
OUTPUT := dist
OUTPUT_COMMIT := 0000000000000000000000000000000000000000
OUTPUT_DATE := 0000-00-00T00:00:00+00:00
OUTPUT_PACKAGE := game_imdb_v1
OUTPUT_VERSION := 0000.00.00-rc
REVIEWDOG_REPORTER := local

.PHONY: tools
tools: ## Tools
	$(GO) mod tidy
	grep _ tools.go | cut -d '"' -f2 | xargs -I{} -n1 -t $(GO) install {}@latest
	$(GO) install github/M2A96/Monopoly.git/internal/cmd/protoc-gen-go-grpc-mock

.PHONY: build
build: ## Build
	$(GO) build \
-ldflags="-X='main.commit=$(OUTPUT_COMMIT)' -X='main.date=$(OUTPUT_DATE)' -X='main.version=$(OUTPUT_VERSION)' -extldflags='-static' -s -w" \
-a \
-buildmode=exe \
\
-o=$(OUTPUT)/ \
-trimpath=true \
./cmd/${OUTPUT_PACKAGE}

.PHONY: add-buf
add-buf: ## Add Buf
	$(GO) install github.com/bufbuild/buf/cmd/buf@latest
	$(GO) install github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking@latest
	$(GO) install github.com/bufbuild/buf/cmd/protoc-gen-buf-lint@latest
	$(GO) install github.com/envoyproxy/protoc-gen-validate/cmd/protoc-gen-validate-go@latest
	$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	$(GO) install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	$(GO) install github/M2A96/Monopoly.git/internal/cmd/protoc-gen-go-grpc-mock
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOPATH)/bin/buf dep update api

.PHONY: buf-breaking
buf-breaking: add-buf ## Buf Breaking
	$(GOPATH)/bin/buf breaking api --against=$(BUF_BREAKING_AGAINST)

.PHONY: buf-format
buf-format: add-buf ## Buf Format
	$(GOPATH)/bin/buf format api --write

.PHONY: buf-generate
buf-generate: add-buf ## Buf Generate
	rm -fr ./gen/
	$(GOPATH)/bin/buf generate api
	## REMEMBER TO CONFIG TEST FILES
	# rm -fr ./test/k6/gen/
	# npm --prefix ./test/k6 install
	# npm --prefix ./test/k6 run buf-generate

.PHONY: buf-lint
buf-lint: add-buf ## Buf
	$(GOPATH)/bin/buf lint api

.PHONY: commitlint
commitlint: ## Commit Lint
	$(GO) install github.com/conventionalcommit/commitlint@latest
	$(GOPATH)/bin/commitlint lint

.PHONY: coverage
coverage: test ## Coverage
	$(GO) install gotest.tools/gotestsum@latest
	$(GO) tool cover -func=./profile/cover.txt
	$(GO) tool cover -func=./profile/cover.txt -o=./profile/cover.txt
	$(GOPATH)/bin/gotestsum --junitfile=./profile/junit.xml -- -count=1 -covermode=atomic -race -tags=$(GO_TAGS) $(GO_PKG_LIST)

.PHONY: doc
doc: ## Documentation
	$(GO) doc -all .

.PHONY: fix
fix: ## Fix
	$(GO) fix $(GO_PKG_LIST)

.PHONY: fmt
fmt: ## Format
	$(GO) fmt $(GO_PKG_LIST)

.PHONY: generate
generate: ## Generate
	$(GO) install go.uber.org/mock/mockgen@latest
	$(GO) install golang.org/x/tools/cmd/stringer@latest
	$(GO) generate $(GO_PKG_LIST)

.PHONY: golangci-lint
golangci-lint: ## GolangCI Lint
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOPATH)/bin/golangci-lint run --build-tags=$(GO_TAGS) --config=./.golangci.yaml $(GO_PKG_LIST)

.PHONY: golines
golines: ## GoLines
	$(GO) install github.com/segmentio/golines@latest
	$(GOPATH)/bin/golines --reformat-tags --write-output .

.PHONY: goreleaser
goreleaser: ## GoReleaser
	$(GO) install github.com/anchore/syft/cmd/syft@latest
	$(GO) install github.com/goreleaser/goreleaser@latest
	$(GO) install github.com/sigstore/cosign/v2/cmd/cosign@latest
	$(GOPATH)/bin/goreleaser release --clean --config=./.goreleaser.yaml --snapshot

.PHONY: govulncheck
govulncheck: ## GoVulncheck
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	$(GOPATH)/bin/govulncheck $(GO_PKG_LIST)

.PHONY: k6
k6: k6-test-build k6-test ## K6

.PHONY: k6-build
k6-build: ## K6 Build
	$(GO) install go.k6.io/xk6/cmd/xk6@latest
	$(GOPATH)/bin/xk6 build latest \
--output $(GOPATH)/bin/ \
--with github.com/grafana/xk6-disruptor \
--with github.com/grafana/xk6-grpc \
--with github.com/grafana/xk6-kubernetes

.PHONY: k6-test
k6-test: k6-build migrate-drop migrate-up ## K6 Test
	$(GOPATH)/bin/k6 run ./test/k6/dist/$(K6_PACKAGE).test.js \
--compatibility-mode=base \
--config=./test/k6/config.json \
--env=ADDRESS=$(K6_ADDRESS) \
--env=BASE_URL=$(K6_BASE_URL) \
--env=SLEEP_DURATION=$(K6_SLEEP_DURATION) \
--env=TESTMAIL_API_KEY=$(K6_TESTMAIL_API_KEY) \
--env=TESTMAIL_NAMESPACE=$(K6_TESTMAIL_NAMESPACE) \
--env=USER_ID=$(K6_USER_ID) \
--summary-export=./test/k6/dist/load-performance.json

.PHONY: k6-test-build
k6-test-build: ## K6 Test Build
	npm --prefix ./test/k6 install
	npm --prefix ./test/k6 run build

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

.PHONY: reviewdog
reviewdog: ## Review Dog
	$(GO) install github.com/reviewdog/reviewdog/cmd/reviewdog@latest
	$(GOPATH)/bin/reviewdog -conf=./.reviewdog.yaml -fail-on-error=true -filter-mode=nofilter -reporter=$(REVIEWDOG_REPORTER)

.PHONY: shfmt
shfmt: ## Shell Formatter
	$(GO) install mvdan.cc/sh/v3/cmd/shfmt@latest
	$(GOPATH)/bin/shfmt --case-indent --indent=2 --write script/*.sh

.PHONY: test
test: ## Test
	mkdir -p profile
	$(GO) test $(GO_PKG_LIST) -count=1 -covermode=atomic -coverprofile=./profile/cover.txt -race -tags=$(GO_TAGS)
