IMAGE_TAG ?= sickrage_example

.PHONY: run docker

run: docker
	@echo "--> Running $(IMAGE_TAG)"
	docker run --rm --name sickrage \
		--publish 8081:8081 $(IMAGE_TAG) \
		--branch=master --update-interval=1h \
		--uid=7000 --gid=7000 \
		SickRage/SickRage sickrage

docker:
	@echo "--> Building base container"
	@cd ../.. && make docker && cd -
	@echo "--> Building $(IMAGE_TAG) Docker image"
	@docker build -t $(IMAGE_TAG) .

help:
	@echo "Available targets:"
	@echo ""
	@echo "  docker"
	@echo "  run (default)"