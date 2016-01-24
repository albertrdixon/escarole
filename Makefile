PROJECT = github.com/albertrdixon/escarole
EXECUTABLE = escarole
IMAGE_TAG ?= escarole_base
PKG = .
LDFLAGS ?= "-s"
PLATFORMS ?= linux darwin

.PHONY: docker release save restore test build install package container clean

release: test package
docker: container

help:
	@echo "Available targets:"
	@echo ""
	@echo "  build"
	@echo "  clean"
	@echo "  container"
	@echo "  docker"
	@echo "  install"
	@echo "  package"
	@echo "  release (default)"
	@echo "  restore"
	@echo "  save"
	@echo "  test"

save:
	godep save ./...

restore:
	godep restore

test:
	@echo "--> Running all tests"
	@echo ""
	@godep go test $(TEST_OPTS) ./...

build:
	@echo "--> Building $(EXECUTABLE) with ldflags '$(LDFLAGS)'"
	@GOOS=linux CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags $(LDFLAGS) -o bin/$(EXECUTABLE)-linux $(PKG)
	@GOOS=darwin CGO_ENABLED=0 godep go build -a -ldflags $(LDFLAGS) -o bin/$(EXECUTABLE)-darwin $(PKG)

install:
	@echo "--> Installing $(EXECUTABLE) with ldflags $(LDFLAGS)"
	@godep go install -ldflags $(LDFLAGS) $(INSTALL)

package: build
	@echo "--> Tar'ing up the binaries"
	@for p in $(PLATFORMS) ; do \
		echo "--> Tar'ing up $$p/amd64 binary" ; \
		test -f bin/$(EXECUTABLE)-$$p && \
		tar czf $(EXECUTABLE)-$$p.tgz bin/$(EXECUTABLE)-$$p ; \
	done

container: build
	@echo "--> Building $(IMAGE_TAG) Docker image"
	@docker build -t $(IMAGE_TAG) .

clean:
	go clean ./...
	rm -rf escarole *.tar.gz
