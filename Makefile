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
BINARY_DARWIN_AMD64=$(BINARY)_darwin
BINARY_DARWIN_ARM64=$(BINARY)_darwin_arm64
BINARY_LINUX_AMD64=$(BINARY)_linux
BINARY_LINUX_ARM64=$(BINARY)_linux_arm64
BINARY_PI=$(BINARY)_pi
BINARY_WINDOWS_AMD64=$(BINARY).exe
BINARY_WINDOWS_ARM64=$(BINARY)_arm64.exe
README=README.md

TESTS=./...
COVERAGE_FILE=coverage.out

BUILD=build/

RESTIC_GEN=$(BUILD)restic-generator
RESTIC_DIR=$(BUILD)restic-
RESTIC_CMD=$(BUILD)restic-commands.json

JSONSCHEMA_DIR=docs/static/jsonschema
CONFIG_REFERENCE_DIR=docs/content/reference

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

all: prepare_test test build
.PHONY: test test-ci coverage
.PHONY: download download-restic-key
.PHONY: build build-mac build-linux build-pi build-windows build-no-selfupdate build-all
.PHONY: generate-config-reference generate-jsonschema generate-install generate-restic
.PHONY: all verify prepare_test prepare_build install clean ramdisk rest-server nightly toc syslog checkdoc

verify:
ifeq ($(wildcard $(GOPATH)/.),)
	@echo "GOPATH not found, please check your go installation"
	exit 1
endif

$(GOBIN)/eget: verify
	@echo "[*] $@"
	GOBIN="$(GOBIN)" $(GOCMD) install -v github.com/zyedidia/eget@v1.3.4

$(GOBIN)/goreleaser: verify $(GOBIN)/eget
	@echo "[*] $@"
	"$(GOBIN)/eget" goreleaser/goreleaser --upgrade-only --to '$(GOBIN)'

$(GOBIN)/github-markdown-toc.go: verify $(GOBIN)/eget
	@echo "[*] $@"
	"$(GOBIN)/eget" ekalinin/github-markdown-toc.go --upgrade-only --file gh-md-toc --to '$(GOBIN)'

$(GOBIN)/mockery: verify $(GOBIN)/eget
	@echo "[*] $@"
	"$(GOBIN)/eget" vektra/mockery --tag v2.53.3 --upgrade-only --to '$(GOBIN)'

$(GOBIN)/golangci-lint: verify $(GOBIN)/eget
	@echo "[*] $@"
	"$(GOBIN)/eget" golangci/golangci-lint --tag v1.64.8 --asset=tar.gz --upgrade-only --to '$(GOBIN)'

$(GOBIN)/hugo: $(GOBIN)/eget
	@echo "[*] $@"
	"$(GOBIN)/eget" gohugoio/hugo --tag 0.145.0 --upgrade-only --asset=extended_0 --to '$(GOBIN)'

prepare_build: verify download
	@echo "[*] $@"

prepare_test: verify download $(GOBIN)/mockery
	@echo "[*] $@"
	find . -path "*/mocks/*" -exec rm {} \;
	"$(GOBIN)/mockery" --config .mockery.yaml

download: verify
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
	$(GOMOD) download

download-restic-key:
	@echo "[*] $@"
	KEY_FILE=$(abspath restic/gpg-key.asc)
	curl https://restic.net/gpg-key-alex.asc > $(KEY_FILE)

install: prepare_build
	@echo "[*] $@"
	GOBIN="$(GOBIN)" \
	$(GOINSTALL) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build: prepare_build
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
	$(GOBUILD) -o $(BINARY) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-no-selfupdate: prepare_build
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
	$(GOBUILD) -o $(BINARY) -v -tags no_self_update -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-mac: prepare_build
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
	GOOS="darwin" GOARCH="amd64" $(GOBUILD) -o $(BINARY_DARWIN_AMD64) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"
	GOPATH="$(GOPATH)" \
	GOOS="darwin" GOARCH="arm64" $(GOBUILD) -o $(BINARY_DARWIN_ARM64) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-linux: prepare_build
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
	GOOS="linux" GOARCH="amd64" $(GOBUILD) -o $(BINARY_LINUX_AMD64) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"
	GOPATH="$(GOPATH)" \
	GOOS="linux" GOARCH="arm64" $(GOBUILD) -o $(BINARY_LINUX_ARM64) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-pi: prepare_build
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
	GOOS="linux" GOARCH="arm" GOARM="6" $(GOBUILD) -o $(BINARY_PI) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-windows: prepare_build
	@echo "[*] $@"
	GOPATH="$(GOPATH)" \
	GOOS="windows" GOARCH="amd64" $(GOBUILD) -o $(BINARY_WINDOWS_AMD64) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"
	GOPATH="$(GOPATH)" \
	GOOS="windows" GOARCH="arm64" $(GOBUILD) -o $(BINARY_WINDOWS_ARM64) -v -ldflags "-X 'main.commit=${BUILD_COMMIT}' -X 'main.date=${BUILD_DATE}' -X 'main.builtBy=make'"

build-all: build-mac build-linux build-pi build-windows

test: prepare_test
	@echo "[*] $@"
	$(GOTEST) $(TESTS)

test-ci: prepare_test
	@echo "[*] $@"
	$(GOTEST) -v -race -short -coverprofile='coverage.out' ./...

coverage:
	@echo "[*] $@"
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TESTS)
	$(GOTOOL) cover -html=$(COVERAGE_FILE)

clean:
	@echo "[*] $@"
	$(GOCLEAN)
	rm -rf $(BINARY) \
	       $(BINARY_DARWIN_AMD64) \
	       $(BINARY_DARWIN_ARM64) \
	       $(BINARY_LINUX_AMD64) \
	       $(BINARY_LINUX_ARM64) \
	       $(BINARY_PI) \
	       $(BINARY_WINDOWS_AMD64) \
	       $(BINARY_WINDOWS_ARM64) \
	       $(COVERAGE_FILE) \
	       restic_*_linux_amd64* \
	       ${BUILD}restic* \
	       ${BUILD}rclone* \
	       dist/*
	find . -path "*/mocks/*" -exec rm {} \;
	restic cache --cleanup

ramdisk: ${TMP_MOUNT}

# Fixed size ramdisk for mac OS X
${TMP_MOUNT_DARWIN}:
	# blocks = 512B so it's creating a 2GB fixed size disk image
	diskutil erasevolume HFS+ RAMDisk `hdiutil attach -nomount ram://4194304`

# Mount tmpfs on linux
${TMP_MOUNT_LINUX}:
	mkdir -p ${TMP_MOUNT_LINUX}
	sudo mount -t tmpfs -o "rw,relatime,size=2097152k,uid=`id -u`,gid=`id -g`" tmpfs ${TMP_MOUNT_LINUX}

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
	GITLAB_TOKEN= goreleaser --snapshot --skip=publish --clean

toc: $(GOBIN)/github-markdown-toc.go
	@echo "[*] $@"
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
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.16.0 --version=0.16.0 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.16.1 --version=0.16.1 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.16.4 --version=0.16.4 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.17.0 --version=0.17.0 --commands $(RESTIC_CMD)
	$(RESTIC_GEN) --install $(RESTIC_DIR)0.18.0 --version=0.18.0 --commands $(RESTIC_CMD)

	cp $(RESTIC_CMD) restic/commands.json

generate-jsonschema: build
	@echo "[*] $@"

	mkdir -p $(JSONSCHEMA_DIR) || echo "$(JSONSCHEMA_DIR) exists"

	$(abspath $(BINARY)) generate --json-schema global > $(JSONSCHEMA_DIR)/config.json

	for config_version in 1 2 ; do \
		$(abspath $(BINARY)) generate --json-schema v$$config_version > $(JSONSCHEMA_DIR)/config-$$config_version.json ; \
		for restic_version in 0.9 0.10 0.11 0.12 0.13 0.14 0.15 0.16 0.17 0.18 ; do \
			name=$$(echo $$restic_version | sed 's/\./-/g') ; \
			$(abspath $(BINARY)) generate --json-schema --version $$restic_version v$$config_version > $(JSONSCHEMA_DIR)/config-$$config_version-restic-$$name.json ; \
		done ; \
	done

generate-config-reference: build
	@echo "[*] $@"

	META_TITLE="Resticprofile configuration reference" \
	META_WEIGHT="6" \
	LAYOUT_NO_HEADLINE="1" \
	LAYOUT_HEADINGS_START="#" \
	LAYOUT_NOTICE_START='{{% notice style="note" %}}' \
	LAYOUT_NOTICE_END="{{% /notice %}}" \
	LAYOUT_HINT_START='{{% notice style="tip" %}}' \
	LAYOUT_HINT_END="{{% /notice %}}" \
	LAYOUT_UPLINK="[go to top](#reference)" \
	$(abspath $(BINARY)) generate --config-reference --to $(CONFIG_REFERENCE_DIR)

.PHONY: documentation
documentation: generate-jsonschema generate-config-reference $(GOBIN)/hugo
	@echo "[*] $@"
	cd docs && hugo --minify

.PHONY: syslog-ng
syslog-ng:
	@echo "[*] $@"
	docker run -d \
		--name=syslog-ng \
		--rm \
		-e PUID=1000 \
		-e PGID=1000 \
		-e TZ=Etc/UTC \
		-p 5514:5514/udp \
		-p 5514:6601/tcp \
		-v $(CURRENT_DIR)/examples/syslog-ng:/config \
		-v $(CURRENT_DIR)/log:/var/log \
		lscr.io/linuxserver/syslog-ng:latest

checkdoc:
	@echo "[*] $@"
	$(GOCMD) run ./config/checkdoc -r docs/content -i changelog.md

.PHONY: checklinks
checklinks:
	@echo "[*] $@"
	muffet --buffer-size=8192 --max-connections-per-host=8 --rate-limit=20 \
	  --exclude="(linux\.die\.net|scoop\.sh|commit)" \
	  --header="User-Agent: Muffet/$$(muffet --version)" \
	  http://127.0.0.1:1313/resticprofile/

.PHONY: lint
lint: $(GOBIN)/golangci-lint
	@echo "[*] $@"
	GOOS=darwin golangci-lint run
	GOOS=linux golangci-lint run
	GOOS=windows golangci-lint run

.PHONY: fix
fix: $(GOBIN)/golangci-lint
	@echo "[*] $@"
	$(GOCMD) mod tidy
	$(GOCMD) fix ./...
	GOOS=darwin golangci-lint run --fix
	GOOS=linux golangci-lint run --fix
	GOOS=windows golangci-lint run --fix

.PHONY: deploy-current
deploy-current: build-linux build-pi
	@echo "[*] $@"
	for server in $$(cat targets_amd64.txt); do \
		echo "Deploying to $$server" ; \
		rsync -avz --progress $(BINARY_LINUX_AMD64) $$server: ; \
		ssh $$server "sudo -S install $(BINARY_LINUX_AMD64) /usr/local/bin/resticprofile" ; \
	done
	for server in $$(cat targets_armv6.txt); do \
		echo "Deploying to $$server" ; \
		rsync -avz --progress $(BINARY_PI) $$server: ; \
		ssh $$server "sudo -S install $(BINARY_PI) /usr/local/bin/resticprofile" ; \
	done
