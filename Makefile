PROJECT = github.com/albertrdixon/escarole
EXECUTABLE = "escarole"
PKG = .
LDFLAGS = "-s"
TEST_COMMAND = godep go test
PLATFORMS = linux darwin
BUILD_ARGS = ""

.PHONY: dep-save dep-restore test test-verbose build install clean

all: test

help:
	@echo "Available targets:"
	@echo ""
	@echo "  dep-save"
	@echo "  dep-restore"
	@echo "  test"
	@echo "  test-verbose"
	@echo "  build"
	@echo "  build-docker"
	@echo "  install"
	@echo "  clean"

dep-save:
	godep save ./...

dep-restore:
	godep restore

test:
	@echo "==> Running all tests"
	@echo ""
	@$(TEST_COMMAND) ./...

test-verbose:
	@echo "==> Running all tests (verbose output)"
	@ echo ""
	@$(TEST_COMMAND) -test.v ./...

build:
	@echo "--> Building $(EXECUTABLE) with ldflags '$(LDFLAGS)'"
	@ GOOS=linux CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags $(LDFLAGS) -o bin/$(EXECUTABLE)-linux $(PKG)
	@ GOOS=darwin CGO_ENABLED=0 godep go build -a -ldflags $(LDFLAGS) -o bin/$(EXECUTABLE)-darwin $(PKG)

install:
	@echo "==> Installing $(EXECUTABLE) with ldflags $(LDFLAGS)"
	@godep go install -ldflags $(LDFLAGS) $(INSTALL)

package: build
	@echo "==> Tar'ing up the binaries"
	@for p in $(PLATFORMS) ; do \
		echo "==> Tar'ing up $$p/amd64 binary" ; \
		test -f bin/$(EXECUTABLE)-$$p && \
		tar czf $(EXECUTABLE)-$$p.tgz bin/$(EXECUTABLE)-$$p ; \
	done

container: build
	@echo "--> Building Docker image"
	@docker build -t escarole .

clean:
	go clean ./...
	rm -rf escarole *.tar.gz
