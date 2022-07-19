GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CURRENT_DIR := $(dir $(MAKEFILE_PATH))
PROJECT_NAME = stand-schedule-policy-controller
DOCKER_IMAGE_COMMIT_SHA=$(shell git show -s --format=%h)
DOCKER_IMAGE_REPO = dodoreg.azurecr.io/${PROJECT_NAME}

BUILD_OS := $(shell uname | sed 's/./\L&/g')
BUILD_ARCH := $(shell uname -m)
ifeq ($(BUILD_ARCH),x86_64)
	BUILD_ARCH = amd64
endif

.PHONY: all
all: help

.PHONY: prepare
prepare: tidy lint

.PHONY: lint
lint:
	golangci-lint run

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: build
build:
	@GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) CGO_ENABLED=0 go build -v -o ./bin/${PROJECT_NAME} ./cmd

.PHONY: build-docker
build-docker: BUILD_OS = linux
build-docker: build
	docker build \
	--progress plain \
	--platform linux/${BUILD_ARCH} \
	--tag "${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_COMMIT_SHA}" \
	--file Dockerfile \
	.

.PHONY: push-docker
push-docker: BUILD_OS = linux
push-docker: build-docker
	docker push "${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_COMMIT_SHA}"

.PHONY: test
test:
	go test -v -timeout 300s --tags=integration ./test/...

.PHONY: run
run: build
	@./bin/${PROJECT_NAME}

.PHONY: run-docker
run-docker: build-docker
	@docker run \
	-it \
	--rm "${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_COMMIT_SHA}"

.PHONY: help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@echo "  ${YELLOW}prepare                   ${RESET} Run all available checks and generators"
	@echo "  ${YELLOW}lint                      ${RESET} Run linters via golangci-lint"
	@echo "  ${YELLOW}tidy                      ${RESET} Run tidy for go module to remove unused dependencies"
	@echo "  ${YELLOW}build                     ${RESET} Build app locally"
	@echo "  ${YELLOW}build-docker              ${RESET} Build app locally and create docker image"
	@echo "  ${YELLOW}push-docker               ${RESET} Build app locally and push docker image"
	@echo "  ${YELLOW}test                      ${RESET} Run tests locally"
	@echo "  ${YELLOW}run                       ${RESET} Run app locally"
	@echo "  ${YELLOW}run-docker                ${RESET} Run app in docker"
