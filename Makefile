GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CURRENT_DIR := $(dir $(MAKEFILE_PATH))
PROJECT_NAME = stand-schedule-policy-controller
DOCKER_IMAGE_COMMIT_SHA=$(shell git show -s --format=%h)
DOCKER_IMAGE_REGISTRY = acrdodo.azurecr.io
DOCKER_IMAGE_REPO = ${DOCKER_IMAGE_REGISTRY}/${PROJECT_NAME}
CONTROLLER_GEN=${GOPATH}/bin/controller-gen
CONTROLLER_GEN_REQ_VERSION := v0.9.1-0.20220629131006-1878064c4cdf
PLUGIN_GIT_TAG := $(shell git tag -l --sort=-creatordate | head -n 1)

BUILD_OS := $(shell uname | sed 's/./\L&/g')
BUILD_ARCH := $(shell uname -m)
ifeq ($(BUILD_ARCH),x86_64)
	BUILD_ARCH = amd64
endif

INTEGRATION_TEST_KIND_CLUSTER_NODE=v1.21.1
INTEGRATION_TEST_CRDS=./crds/StandSchedulePolicy.yaml
INTEGRATION_TEST_KIND_CLUSTER_CONFIG=$(shell pwd)/bin/kubeconfig.yaml

.PHONY: all
all: help

.PHONY: prepare
prepare: tidy lint ## Run all available checks and generators

.PHONY: codegen
codegen: ## Run code generators for CRDs
	./hack/run-codegen.sh

.PHONY: lint
lint: ## Run linters via golangci-lint
	golangci-lint run

.PHONY: tidy
tidy: ## Run tidy for go module to remove unused dependencies
	go mod tidy -v

.PHONY: build
build: build-controller build-plugin ## Build controller and plugin locally

.PHONY: build-controller
build-controller: ## Build controller locally
	@GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) CGO_ENABLED=0 go build -v -o ./bin/${PROJECT_NAME} ./cmd/controller

.PHONY: build-plugin
build-plugin: ## Build plugin locally
	@GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) CGO_ENABLED=0 go build -v -o ./bin/${PROJECT_NAME} ./cmd/plugin

.PHONY: build-docker
build-docker: BUILD_OS = linux
build-docker: build-controller ## Build controller locally and create docker image
	docker build \
	--progress plain \
	--platform linux/${BUILD_ARCH} \
	--tag "${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_COMMIT_SHA}" \
	--file Dockerfile \
	.

.PHONY: push-docker
push-docker: BUILD_OS = linux
push-docker: build-docker ## Build controller locally and push docker image
	docker push "${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_COMMIT_SHA}"

.PHONY: test
test: test-unit test-integration ## Run all tests

.PHONY: test-unit
test-unit: ## Run unit tests
	go test \
	-v \
	-timeout 30s \
	./...

.PHONY: test-integration-setup
test-integration-setup: ## Setup environment for integration tests
	kind create cluster \
	--name kind \
	--image=kindest/node:$(INTEGRATION_TEST_KIND_CLUSTER_NODE) \
	--kubeconfig "${INTEGRATION_TEST_KIND_CLUSTER_CONFIG}"
	kubectl apply \
	--kubeconfig "${INTEGRATION_TEST_KIND_CLUSTER_CONFIG}" \
	--filename "${INTEGRATION_TEST_CRDS}"
	sleep 20s

.PHONY: test-integration-cleanup
test-integration-cleanup: ## Cleanup environment for integration tests
	kind delete cluster \
	--name kind

.PHONY: test-integration
test-integration: ## Run all integration tests
	TEST_KUBECONFIG_PATH="${INTEGRATION_TEST_KIND_CLUSTER_CONFIG}" go test \
	-v \
	-timeout 600s \
	--tags=integration \
	-parallel 1 \
	./test/...

.PHONY: run-controller
run-controller: build-controller ## Run controller locally
	@./bin/${PROJECT_NAME}

.PHONY: run-docker
run-docker: build-docker ## Run controller in docker
	@docker run \
	-it \
	--rm "${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_COMMIT_SHA}"

.PHONY: plugin-template
plugin-template: ## Build krew spec for plugin based on .krew.yaml
	docker run \
		--rm \
		-v ${CURRENT_DIR}/.krew.yaml:/tmp/template-file.yaml \
		rajatjindal/krew-release-bot:v0.0.43 \
		krew-release-bot template \
		--tag "${PLUGIN_GIT_TAG}" \
		--template-file /tmp/template-file.yaml

.PHONY: help
help: ## Shows the available commands
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  ${YELLOW}%-30s${RESET} %s\n", $$1, $$2}'