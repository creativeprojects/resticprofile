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
GOMOD=$(GOCMD) mod
GOPATH?=`$(GOCMD) env GOPATH`

BINARY=resticprofile
BINARY_DARWIN=$(BINARY)_darwin
BINARY_LINUX=$(BINARY)_linux
BINARY_PI=$(BINARY)_pi
BINARY_WINDOWS=$(BINARY).exe
README=README.md

TESTS=./...
COVERAGE_FILE=coverage.out

BUILD=build/

BUILD_DATE=`date`
BUILD_COMMIT=`git rev-parse HEAD`

CURRENT_DIR = $(shell pwd)

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

TOC_START=<\!--ts-->
TOC_END=<\!--te-->
TOC_PATH=toc.md

all: download test build
.PHONY: all download test test-ci build install build-mac build-linux build-windows build-all coverage clean ramdisk passphrase rest-server nightly toc release-snapshot generate-install

download:
	@echo "[*] $@"
	@$(GOMOD) download

build: download
	@echo "[*] $@"
	$(GOBUILD) -o $(BINARY) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

install: download
	@echo "[*] $@"
	$(GOINSTALL) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-mac: download
	@echo "[*] $@"
	GOOS="darwin" GOARCH="amd64" $(GOBUILD) -o $(BINARY_DARWIN) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-linux: download
	@echo "[*] $@"
	GOOS="linux" GOARCH="amd64" $(GOBUILD) -o $(BINARY_LINUX) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-pi: download
	@echo "[*] $@"
	GOOS="linux" GOARCH="arm" GOARM="6" $(GOBUILD) -o $(BINARY_PI) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-windows: download
	@echo "[*] $@"
	GOOS="windows" GOARCH="amd64" $(GOBUILD) -o $(BINARY_WINDOWS) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-all: build-mac build-linux build-pi build-windows

test: download
	@echo "[*] $@"
	$(GOTEST) -v $(TESTS)

test-ci: download
	@echo "[*] $@"
	$(GOTEST) -v -short ./...
	$(GOTEST) -v ./priority

coverage:
	@echo "[*] $@"
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TESTS)
	$(GOTOOL) cover -html=$(COVERAGE_FILE)

clean:
	@echo "[*] $@"
	$(GOCLEAN)
	rm -rf $(BINARY) $(BINARY_DARWIN) $(BINARY_LINUX) $(BINARY_PI) $(BINARY_WINDOWS) $(COVERAGE_FILE) restic_*_linux_amd64* ${BUILD}restic* dist/*
	restic cache --cleanup

release-snapshot:
	@echo "[*] $@"
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

rest-server:
	@echo "[*] $@"
	REST_IMAGE=restic/rest-server
	REST_CONTAINER=rest_server
	REST_DATA=/tmp/restic
	REST_OPTIONS=""

	docker pull ${REST_IMAGE}
	docker run -d -p 8000:8000 -v ${REST_DATA}:/data --name ${REST_CONTAINER} --restart always -e "OPTIONS=${REST_OPTIONS}" ${REST_IMAGE}

nightly:
	@echo "[*] $@"
	$(GOINSTALL) github.com/goreleaser/goreleaser@latest
	goreleaser --snapshot --skip-publish --rm-dist

toc:
	@echo "[*] $@"
	$(GOINSTALL) github.com/ekalinin/github-markdown-toc.go/cmd/gh-md-toc@latest
	cat ${README} | gh-md-toc --hide-footer > ${TOC_PATH}
	sed -i ".1" "/${TOC_START}/,/${TOC_END}/{//!d;}" "${README}"
	sed -i ".2" "/${TOC_START}/r ${TOC_PATH}" "${README}"
	rm ${README}.1 ${README}.2 ${TOC_PATH}

generate-install:
	@echo "[*] $@"
	godownloader .godownloader.yml -r creativeprojects/resticprofile -o install.sh

syslog:
	@echo "[*] $@"
	docker run -d \
		--name=rsyslogd \
		--rm \
		-p 5514:514/udp \
		-p 5514:514/tcp \
		-v $(CURRENT_DIR)/examples/rsyslogd.conf:/etc/rsyslog.d/listen.conf \
		instantlinux/rsyslogd:latest
