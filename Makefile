# Copyright © 2022-2024 Christian Fritz <mail@chr-fritz.de>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL = /usr/bin/env bash -o pipefail -o errexit -o nounset
NAME := knx-exporter
ORG := chr-fritz
ROOT_PACKAGE := github.com/chr-fritz/knx-exporter

TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(TAG:v%=%)
ifneq ($(COMMIT), $(TAG_COMMIT))
    VERSION := $(VERSION)-next
endif


REVISION   := $(shell git rev-parse --short HEAD 2> /dev/null  || echo 'unknown')
BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell git show -s --format=%ct)

PACKAGE_DIRS := $(shell go list ./...)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./bin
REPORTS_DIR ?= ./reports

BUILDFLAGS := -ldflags \
  " -X '$(ROOT_PACKAGE)/version.Version=$(VERSION)'\
    -X '$(ROOT_PACKAGE)/version.Revision=$(REVISION)'\
    -X '$(ROOT_PACKAGE)/version.Branch=$(BRANCH)'\
    -X '$(ROOT_PACKAGE)/version.CommitDate=$(BUILD_DATE)'\
    -s -w -extldflags '-static'"

.PHONY: all
all: test $(GOOS)-build
	@echo "SUCCESS"

.PHONY: ci
ci: ci-check

.PHONY: ci-check
ci-check: tidy generate imports vet test

.PHONY: build
build:
	CGO_ENABLED=0 GOARCH=amd64 go build $(BUILDFLAGS) -o $(BUILD_DIR)/$(NAME) $(ROOT_PACKAGE)

.PHONY: debug
debug:
	CGO_ENABLED=0 GOARCH=amd64 go build -gcflags "all=-N -l" -o $(BUILD_DIR)/$(NAME)-debug $(ROOT_PACKAGE)
	dlv --listen=:2345 --headless=true --api-version=2 exec $(BUILD_DIR)/$(NAME)-debug run

.PHONY: imports
imports:
	find . -type f -name '*.go' ! -name '*_mocks.go' -print0 | xargs -0 goimports -w -l

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: darwin-build
darwin-build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build $(BUILDFLAGS) -o $(BUILD_DIR)/$(NAME)-darwin $(ROOT_PACKAGE)

.PHONY: linux-build
linux-build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build $(BUILDFLAGS) -o $(BUILD_DIR)/$(NAME)-linux $(ROOT_PACKAGE)

.PHONY: windows-build
windows-build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build $(BUILDFLAGS) -o $(BUILD_DIR)/$(NAME)-windows.exe $(ROOT_PACKAGE)

.PHONY: test
test: generate
	mkdir -p $(REPORTS_DIR)
	go test $(PACKAGE_DIRS) -coverprofile=$(REPORTS_DIR)/coverage.out -v $(PACKAGE_DIRS) | tee >(go tool test2json > $(REPORTS_DIR)/tests.json)

.PHONY: test-race
test-race: generate
	mkdir -p $(REPORTS_DIR)
	go test -race $(PACKAGE_DIRS) -coverprofile=$(REPORTS_DIR)/coverage.out -v $(PACKAGE_DIRS) | tee >(go tool test2json > $(REPORTS_DIR)/tests.json)

.PHONY: cross
cross: darwin-build linux-build windows-build

.PHONY: vet
vet:
	mkdir -p $(REPORTS_DIR)
	go vet -v $(PACKAGE_DIRS) 2> >(tee $(REPORTS_DIR)/vet.out) || true

.PHONY: lint
lint:
	mkdir -p $(REPORTS_DIR)
	# GOGC default is 100, but we need more aggressive GC to not consume too much memory
	# might not be necessary in future versions of golangci-lint
	# https://github.com/golangci/golangci-lint/issues/483
	GOGC=20 golangci-lint run --disable=typecheck --deadline=5m --out-format checkstyle > $(REPORTS_DIR)/lint.xml || true

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -rf $(REPORTS_DIR)

.PHONY: buildDeps
buildDeps:
	go mod download
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: completions
completions:
	rm -rf completions
	mkdir completions
	for sh in bash zsh fish ps1; do go run main.go completion "$$sh" >"completions/$(NAME).$$sh"; done

.PHONY: sonarcloud-version
sonarcloud-version:
	echo "sonar.projectVersion=$(VERSION)" >> sonar-project.properties
