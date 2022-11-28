PROJECT_ID := poc-gcp-ecomm
GCP_REGION := us-central1
SERVICE_NAME := order-service
RUNTIME := go119

.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	echo "make would usually be run from the parent directory rather than here!\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: gomod test ## Build the gRPC service container locally after running unit tests
	export PROJECT_ID=$(PROJECT_ID); \
	gcloud builds submit --config=cloudbuild.yaml .

.PHONY: deploy
deploy: ## Deploy the the latest gRPC service container from the artifact repository
	gcloud run deploy $(SERVICE_NAME) --image us-central1-docker.pkg.dev/$(PROJECT_ID)/gcr-artifacts/$(SERVICE_NAME):latest --region $(GCP_REGION) --use-http2 --no-allow-unauthenticated

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
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/cart
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/pb
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/types
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/util
	go mod tidy
