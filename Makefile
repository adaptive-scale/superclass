# Image URL to use all building/pushing image targets
REGISTRY ?= ghcr.io
REPOSITORY ?= adaptive-scale/superclass
TAG ?= latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Architectures to build for
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build the binary
	go build -o bin/superclass main.go

.PHONY: run
run: fmt vet ## Run from source
	go run ./main.go

##@ Docker

.PHONY: docker-build
docker-build: ## Build the docker image
	docker build -t $(REGISTRY)/$(REPOSITORY):$(TAG) .

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(REGISTRY)/$(REPOSITORY):$(TAG)

.PHONY: docker-buildx-builder
docker-buildx-builder: ## Create a buildx builder for multi-arch builds
	docker buildx create --name superclass-builder --use || true
	docker buildx inspect --bootstrap

.PHONY: docker-buildx
docker-buildx: docker-buildx-builder ## Build and push multi-arch docker images
	@echo "Building and pushing multi-arch images for $(PLATFORMS)"
	docker buildx build \
		--platform $(PLATFORMS) \
		--tag $(REGISTRY)/$(REPOSITORY):$(TAG) \
		--push \
		.

.PHONY: docker-login
docker-login: ## Login to GitHub Container Registry
	@echo "Logging in to GitHub Container Registry"
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "GITHUB_TOKEN is not set. Please set it before running this command."; \
		exit 1; \
	fi
	@echo $(GITHUB_TOKEN) | docker login $(REGISTRY) -u $(GITHUB_USER) --password-stdin

##@ CI/CD

.PHONY: ci
ci: test docker-buildx ## Run all CI steps

.PHONY: release
release: ## Create and push a new release
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION is not set. Please set it before running this command."; \
		exit 1; \
	fi
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	$(MAKE) docker-buildx TAG=$(VERSION)
	$(MAKE) docker-buildx TAG=latest

##@ Clean

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f cover.out 