BUILD_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo no-commit-id)
IMAGE_NAME := anubhavmishra/vault-kubernetes-authenticator

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build docker image
	docker build -t $(IMAGE_NAME):0.2.0 .
	docker tag $(IMAGE_NAME):0.2.0 $(IMAGE_NAME):latest

push: ## docker push the service images tagged 'latest' & 'BUILD_ID'
	docker push $(IMAGE_NAME):0.2.0
	docker push $(IMAGE_NAME):latest

run: ## run docker image locally
	docker run -it $(IMAGE_NAME):latest