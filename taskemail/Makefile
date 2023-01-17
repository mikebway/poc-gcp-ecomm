# Project Settings
PROJECT_ID := poc-gcp-ecomm
GCP_REGION := us-central1

# Function configuration
ENTRY_POINT := FulfillTask
RUNTIME := go119

# Pub/Sub topic name to subscribe to
PUBSUB_TOPIC := ecomm-task

# Name/ID to give to Pub/Sub subscription
SUBSCRIPTION_ID := ${FUNCTION_NAME}


.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	echo "make would usually be run from the parent directory rather than here!\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: gomod compile ## Cloud Function builds do nothing locally other than ensure that go.mod is up to date and that the code compiles

.PHONY: deploy
deploy: gomod ## Deploy the the email sending Cloud Function multiple times with different names and environment variable settings
	gcloud functions deploy task-gy-man --set-env-vars FULFILL_OPERATION=manufacture-gold-yoyo \
     --gen2 --region $(GCP_REGION) --runtime $(RUNTIME) \
     --entry-point=$(ENTRY_POINT) --trigger-http --ingress-settings=all
	gcloud functions deploy task-py-up --set-env-vars FULFILL_OPERATION=upsell-to-gold-yoyo \
     --gen2 --region $(GCP_REGION) --runtime $(RUNTIME) \
     --entry-point=$(ENTRY_POINT) --trigger-http --ingress-settings=all
	gcloud functions deploy task-sf --set-env-vars FULFILL_OPERATION=salesforce-case \
     --gen2 --region $(GCP_REGION) --runtime $(RUNTIME) \
     --entry-point=$(ENTRY_POINT) --trigger-http --ingress-settings=all
	gcloud functions deploy task-ship --set-env-vars FULFILL_OPERATION=ship-product \
     --gen2 --region $(GCP_REGION) --runtime $(RUNTIME) \
     --entry-point=$(ENTRY_POINT) --trigger-http --ingress-settings=all

.PHONY: test
test: compile ## Run the unit tests locally
	go test ./... -coverprofile cover.out -race; \
   	go tool cover -func cover.out

.PHONY: compile
compile: ## Compile the Go code locally
	go build

.PHONY: gomod
gomod: ## Ensure that monorepo pseudo-versions are up to date with latest github commit
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/fulfillment
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/pb
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/types
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/util
	go mod tidy
