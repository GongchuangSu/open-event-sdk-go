PROJECT_NAME := openevent
CLI_PKG      := ./cmd/openevent
OUTPUT_DIR   := output
BINARY       := $(OUTPUT_DIR)/$(PROJECT_NAME)

VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE         ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GO           := go
GOFLAGS      :=
CGO_ENABLED  ?= 0
LDFLAGS      := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

.PHONY: all build run test test-race test-cover lint vet clean install help

all: lint test build ## 默认: lint + test + build

## ---- 构建 ----

build: ## 构建 CLI 二进制到 output/
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(BINARY) $(CLI_PKG)
	@echo "✔ $(BINARY) ($(VERSION))"

run: ## 通过 go run 直接运行 CLI（传递 ARGS，如 make run ARGS="listen --help"）
	$(GO) run $(GOFLAGS) -ldflags '$(LDFLAGS)' $(CLI_PKG) $(ARGS)

install: ## 安装到 $GOPATH/bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) install $(GOFLAGS) -ldflags '$(LDFLAGS)' $(CLI_PKG)
	@echo "✔ $(PROJECT_NAME) installed to $$(go env GOPATH)/bin"

## ---- 测试 ----

test: ## 运行全部单元测试
	$(GO) test ./...

test-race: ## 运行测试（含竞态检测）
	$(GO) test -race ./...

test-cover: ## 运行测试并生成覆盖率报告
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	@echo "---"
	@echo "HTML 报告: go tool cover -html=coverage.out -o coverage.html"

## ---- 代码质量 ----

lint: vet ## 代码检查

vet: ## go vet 静态分析
	$(GO) vet ./...

## ---- 清理 ----

clean: ## 清理构建产物
	rm -rf $(OUTPUT_DIR) coverage.out coverage.html

## ---- 帮助 ----

help: ## 显示所有可用目标
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'
