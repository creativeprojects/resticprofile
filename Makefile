# 
# Makefile for resticprofile
# 
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
GOPATH?=`$(GOCMD) env GOPATH`

BINARY=resticprofile
BINARY_DARWIN=$(BINARY)_darwin
BINARY_LINUX=$(BINARY)_linux
BINARY_WINDOWS=$(BINARY).exe

TESTS=./...
COVERAGE_FILE=coverage.out

BUILD=build/
RESTIC_VERSION=0.9.6
GO_VERSION=1.14

.PHONY: all test test-ci build build-mac build-linux build-windows build-all coverage clean test-docker build-docker ramdisk

all: test build

build:
		$(GOBUILD) -o $(BINARY) -v

build-mac:
		GOOS="darwin" GOARCH="amd64" $(GOBUILD) -o $(BINARY_DARWIN) -v

build-linux:
		GOOS="linux" GOARCH="amd64" $(GOBUILD) -o $(BINARY_LINUX) -v

build-windows:
		GOOS="windows" GOARCH="amd64" $(GOBUILD) -o $(BINARY_WINDOWS) -v

build-all: build-mac build-linux build-windows

test:
		$(GOTEST) -v $(TESTS)

test-ci:
		$(GOTEST) -v -short ./...
		$(GOTEST) -v ./priority

coverage:
		$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TESTS)
		$(GOTOOL) cover -html=$(COVERAGE_FILE)

clean:
		$(GOCLEAN)
		rm -f $(BINARY) $(BINARY_DARWIN) $(BINARY_LINUX) $(BINARY_WINDOWS) $(COVERAGE_FILE) restic_*_linux_amd64* ${BUILD}restic*

test-docker:
		docker run --rm -v "${GOPATH}":/go -w /go/src/creativeprojects/resticprofile golang:${GO_VERSION} $(GOTEST) -v $(TESTS)

build-docker: clean
		CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GOBUILD) -v -o ${BUILD}$(BINARY) .
		curl -LO https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/restic_${RESTIC_VERSION}_linux_amd64.bz2
		bunzip2 restic_${RESTIC_VERSION}_linux_amd64.bz2
		mv restic_${RESTIC_VERSION}_linux_amd64 ${BUILD}restic
		chmod +x ${BUILD}restic
		cd ${BUILD}; docker build --pull --tag creativeprojects/resticprofile .

ramdisk: /Volumes/RAMDisk

/Volumes/RAMDisk:
		diskutil erasevolume HFS+ RAMDisk `hdiutil attach -nomount ram://4194304`
