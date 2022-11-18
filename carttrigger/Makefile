PROJECT_ID := poc-gcp-ecomm
GCP_REGION := us-central1
TRIGGER_NAME := UpdateTrigger
RUNTIME := go119

.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	echo "make would usually be run from the parent directory rather than here!\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: gomod test ## Build the gRPC service container locally after running unit tests

.PHONY: deploy
deploy: build ## Deploy the the trigger Cloud Function
	gcloud functions deploy $(TRIGGER_NAME) --region $(GCP_REGION) --runtime $(RUNTIME) \
	 --trigger-event providers/cloud.firestore/eventTypes/document.write \
	 --trigger-resource "projects/$(PROJECT_ID)/databases/(default)/documents/carts/{id}"

.PHONY: run
run: compile ## Run the gRPC server locally
	go run

.PHONY: test
test: compile ## Run the unit tests locally
	go test ./... -coverprofile cover.out -race; \
   	go tool cover -func cover.out

.PHONY: compile
compile: ## Compile the Go code locally
	go build

.PHONY: gomod
gomod: ## Ensure that monorepo pseudo-versions are up to date with latest github commit
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/pb -droprequire=github.com/mikebway/poc-gcp-ecomm/types
	go mod tidy
