GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get

BINARY=resticprofile

TESTS=./...
COVERAGE_FILE=coverage.out

BUILD=build/
RESTIC_VERSION=0.9.6

.PHONY: all test build coverage clean docker

all: test build

build:
		$(GOBUILD) -o $(BINARY) -v

test:
		$(GOTEST) -v $(TESTS)

coverage:
		$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TESTS)
		$(GOTOOL) cover -html=$(COVERAGE_FILE)

clean:
		$(GOCLEAN)
		rm -f $(BINARY) $(COVERAGE_FILE) restic_*_linux_amd64*

docker: clean
		CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GOBUILD) -v -o ${BUILD}$(BINARY) .
		curl -LO https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/restic_${RESTIC_VERSION}_linux_amd64.bz2
		bunzip2 restic_${RESTIC_VERSION}_linux_amd64.bz2
		mv restic_${RESTIC_VERSION}_linux_amd64 ${BUILD}restic
		chmod +x ${BUILD}restic
		cd ${BUILD}; docker build --pull --tag creativeprojects/resticprofile .
