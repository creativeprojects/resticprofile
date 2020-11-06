# 
# Makefile for resticprofile
# 
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
GOPATH?=`$(GOCMD) env GOPATH`

BINARY=resticprofile
BINARY_DARWIN=$(BINARY)_darwin
BINARY_LINUX=$(BINARY)_linux
BINARY_PI=$(BINARY)_pi
BINARY_WINDOWS=$(BINARY).exe

TESTS=./...
COVERAGE_FILE=coverage.out

BUILD=build/
RESTIC_VERSION=0.9.6
GO_VERSION=1.14

BUILD_DATE=`date`
BUILD_COMMIT=`git rev-parse HEAD`

TMP_MOUNT_LINUX=/tmp/backup
TMP_MOUNT_DARWIN=/Volumes/RAMDisk
TMP_MOUNT=
UNAME := $(shell uname -s)
ifeq ($(UNAME),Linux)
	TMP_MOUNT=${TMP_MOUNT_LINUX}
endif
ifeq ($(UNAME),Darwin)
	TMP_MOUNT=${TMP_MOUNT_DARWIN}
endif

.PHONY: all test test-ci build install build-mac build-linux build-windows build-all coverage clean test-docker build-docker ramdisk passphrase rest-server nightly toc staticcheck release-snapshot generate-install

all: test build

build:
		$(GOBUILD) -o $(BINARY) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

install:
		$(GOINSTALL) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-mac:
		GOOS="darwin" GOARCH="amd64" $(GOBUILD) -o $(BINARY_DARWIN) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-linux:
		GOOS="linux" GOARCH="amd64" $(GOBUILD) -o $(BINARY_LINUX) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-pi:
		GOOS="linux" GOARCH="arm" GOARM="6" $(GOBUILD) -o $(BINARY_PI) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-windows:
		GOOS="windows" GOARCH="amd64" $(GOBUILD) -o $(BINARY_WINDOWS) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

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
		rm -rf $(BINARY) $(BINARY_DARWIN) $(BINARY_LINUX) $(BINARY_PI) $(BINARY_WINDOWS) $(COVERAGE_FILE) restic_*_linux_amd64* ${BUILD}restic* dist/*
		restic cache --cleanup

test-docker:
		docker run --rm -v "${GOPATH}":/go -w /go/src/creativeprojects/resticprofile golang:${GO_VERSION} $(GOTEST) -v $(TESTS)

build-docker: clean
		CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GOBUILD) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'" -o ${BUILD}$(BINARY) .
		curl -LO https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/restic_${RESTIC_VERSION}_linux_amd64.bz2
		bunzip2 restic_${RESTIC_VERSION}_linux_amd64.bz2
		mv restic_${RESTIC_VERSION}_linux_amd64 ${BUILD}restic
		chmod +x ${BUILD}restic
		cd ${BUILD}; docker build --pull --tag creativeprojects/resticprofile .

release-snapshot:
		goreleaser build --snapshot --config .goreleaser.yml --rm-dist

ramdisk: ${TMP_MOUNT}

# Fixed size ramdisk for mac OS X
${TMP_MOUNT_DARWIN}:
		# blocks = 512B so it's creating a 2GB fixed size disk image
		diskutil erasevolume HFS+ RAMDisk `hdiutil attach -nomount ram://4194304`

# Mount tmpfs on linux
${TMP_MOUNT_LINUX}:
		mkdir -p ${TMP_MOUNT_LINUX}
		sudo mount -t tmpfs -o "rw,relatime,size=2097152k" tmpfs ${TMP_MOUNT_LINUX}

passphrase:
		head -c 1024 /dev/urandom | base64

rest-server:
	REST_IMAGE=restic/rest-server
	REST_CONTAINER=rest_server
	REST_DATA=/tmp/restic
	REST_OPTIONS=""

	docker pull ${REST_IMAGE}
	docker run -d -p 8000:8000 -v ${REST_DATA}:/data --name ${REST_CONTAINER} --restart always -e "OPTIONS=${REST_OPTIONS}" ${REST_IMAGE}

nightly:
	go install github.com/goreleaser/goreleaser
	go mod tidy
	goreleaser --snapshot --skip-publish --rm-dist

toc:
	go install github.com/ekalinin/github-markdown-toc.go
	go mod tidy
	cat README.md | github-markdown-toc.go --hide-footer

staticcheck:
	go get -u honnef.co/go/tools/cmd/staticcheck
	go mod tidy
	go run honnef.co/go/tools/cmd/staticcheck ./...

generate-install:
	godownloader .goreleaser.yml -r creativeprojects/resticprofile -o install.sh
