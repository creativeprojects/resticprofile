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
GOPATH=$(shell $(GOCMD) env GOPATH)
GOBIN=$(shell $(GOCMD) env GOBIN)

ifeq ($(GOBIN),)
	GOBIN := $(GOPATH)/bin
endif

BINARY=resticprofile
BINARY_DARWIN=$(BINARY)_darwin
BINARY_LINUX=$(BINARY)_linux
BINARY_PI=$(BINARY)_pi
BINARY_WINDOWS=$(BINARY).exe
README=README.md

TESTS=./...
COVERAGE_FILE=coverage.out

BUILD=build/

RESTIC_GEN=$(BUILD)restic-generator
RESTIC_DIR=$(BUILD)restic-
RESTIC_CMD=$(BUILD)restic-commands.json

JSONSCHEMA_DIR=docs/static/jsonschema
CONFIG_REFERENCE_DIR=docs/content/configuration/reference

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

all: prepare test build
.PHONY: all download test test-ci build install build-mac build-linux build-windows build-all coverage clean ramdisk rest-server nightly toc generate-install syslog checkdoc

$(GOBIN)/eget:
	@echo "[*] $@"
	go install -v github.com/zyedidia/eget@latest

$(GOBIN)/goreleaser: $(GOBIN)/eget
	@echo "[*] $@"
	eget goreleaser/goreleaser --to $(GOBIN)

.PHONY: prepare
prepare: download
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
 	$(GOCMD) generate ./...

download:
	@echo "[*] $@"
	@$(GOMOD) download

download-restic-key:
	@echo "[*] $@"
	KEY_FILE=$(abspath restic/gpg-key.asc)
	curl https://restic.net/gpg-key-alex.asc > $(KEY_FILE)

install: prepare
	@echo "[*] $@"
	$(GOINSTALL) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build: prepare
	@echo "[*] $@"
	$(GOBUILD) -o $(BINARY) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-no-selfupdate: prepare
	@echo "[*] $@"
	$(GOBUILD) -o $(BINARY) -v -tags no_self_update -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-mac: prepare
	@echo "[*] $@"
	GOOS="darwin" GOARCH="amd64" $(GOBUILD) -o $(BINARY_DARWIN) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-linux: prepare
	@echo "[*] $@"
	GOOS="linux" GOARCH="amd64" $(GOBUILD) -o $(BINARY_LINUX) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-pi: prepare
	@echo "[*] $@"
	GOOS="linux" GOARCH="arm" GOARM="6" $(GOBUILD) -o $(BINARY_PI) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-windows: prepare
	@echo "[*] $@"
	GOOS="windows" GOARCH="amd64" $(GOBUILD) -o $(BINARY_WINDOWS) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-all: build-mac build-linux build-pi build-windows

test: prepare
	@echo "[*] $@"
	$(GOTEST) -v $(TESTS)

test-ci: prepare
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

nightly: $(GOBIN)/goreleaser
	@echo "[*] $@"
	GITLAB_TOKEN= goreleaser --snapshot --skip-publish --clean --debug

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

generate-restic:
	@echo "[*] $@"
	$(GOBUILD) -o $(RESTIC_GEN) $(abspath restic/generator)

	rm $(RESTIC_CMD) || echo "clean $(RESTIC_CMD)"

	$(RESTIC_GEN) --install $(RESTIC_DIR)0.9.4 --version=0.9.4 --base-version --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.10.0 --version=0.10.0 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.11.0 --version=0.11.0 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.12.0 --version=0.12.0 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.13.0 --version=0.13.0 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.14.0 --version=0.14.0 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.15.0 --version=0.15.0 --commands $(RESTIC_CMD)

	cp $(RESTIC_CMD) restic/commands.json

generate-jsonschema: build
	@echo "[*] $@"

	mkdir -p $(JSONSCHEMA_DIR) || echo "$(JSONSCHEMA_DIR) exists"

	$(abspath $(BINARY)) generate --json-schema v1 > $(JSONSCHEMA_DIR)/config-1.json
	$(abspath $(BINARY)) generate --json-schema v2 > $(JSONSCHEMA_DIR)/config-2.json
	$(abspath $(BINARY)) generate --json-schema --version 0.9 v1 > $(JSONSCHEMA_DIR)/config-1-restic-0-9.json
	$(abspath $(BINARY)) generate --json-schema --version 0.9 v2 > $(JSONSCHEMA_DIR)/config-2-restic-0-9.json
	$(abspath $(BINARY)) generate --json-schema --version 0.10 v1 > $(JSONSCHEMA_DIR)/config-1-restic-0-10.json
	$(abspath $(BINARY)) generate --json-schema --version 0.10 v2 > $(JSONSCHEMA_DIR)/config-2-restic-0-10.json
	$(abspath $(BINARY)) generate --json-schema --version 0.11 v1 > $(JSONSCHEMA_DIR)/config-1-restic-0-11.json
	$(abspath $(BINARY)) generate --json-schema --version 0.11 v2 > $(JSONSCHEMA_DIR)/config-2-restic-0-11.json
	$(abspath $(BINARY)) generate --json-schema --version 0.12 v1 > $(JSONSCHEMA_DIR)/config-1-restic-0-12.json
	$(abspath $(BINARY)) generate --json-schema --version 0.12 v2 > $(JSONSCHEMA_DIR)/config-2-restic-0-12.json
	$(abspath $(BINARY)) generate --json-schema --version 0.13 v1 > $(JSONSCHEMA_DIR)/config-1-restic-0-13.json
	$(abspath $(BINARY)) generate --json-schema --version 0.13 v2 > $(JSONSCHEMA_DIR)/config-2-restic-0-13.json
	$(abspath $(BINARY)) generate --json-schema --version 0.14 v1 > $(JSONSCHEMA_DIR)/config-1-restic-0-14.json
	$(abspath $(BINARY)) generate --json-schema --version 0.14 v2 > $(JSONSCHEMA_DIR)/config-2-restic-0-14.json
	$(abspath $(BINARY)) generate --json-schema --version 0.15 v1 > $(JSONSCHEMA_DIR)/config-1-restic-0-15.json
	$(abspath $(BINARY)) generate --json-schema --version 0.15 v2 > $(JSONSCHEMA_DIR)/config-2-restic-0-15.json

generate-config-reference: build
	@echo "[*] $@"

	META_TITLE="Reference" \
	META_WEIGHT="50" \
	LAYOUT_NO_HEADLINE="1" \
	LAYOUT_HEADINGS_START="#" \
	LAYOUT_NOTICE_START="{{% notice note %}}" \
	LAYOUT_NOTICE_END="{{% /notice %}}" \
	LAYOUT_HINT_START="{{% notice hint %}}" \
	LAYOUT_HINT_END="{{% /notice %}}" \
	$(abspath $(BINARY)) generate --config-reference > $(CONFIG_REFERENCE_DIR)/index.md

syslog:
	@echo "[*] $@"
	docker run -d \
		--name=rsyslogd \
		--rm \
		-p 5514:514/udp \
		-p 5514:514/tcp \
		-v $(CURRENT_DIR)/examples/rsyslogd.conf:/etc/rsyslog.d/listen.conf \
		instantlinux/rsyslogd:latest

checkdoc:
	go run ./config/checkdoc -r docs/content
