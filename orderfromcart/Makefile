# Project Settings
PROJECT_ID := poc-gcp-ecomm
GCP_REGION := us-central1

# Function configuration
FUNCTION_NAME := order-from-cart
ENTRY_POINT := OrderFromCart
RUNTIME := go119

# Pub/Sub topic name to subscribe to
SHOPPING_CART_TOPIC := ecomm-cart

# Name/ID to give to Pub/Sub subscription
SUBSCRIPTION_ID := ${FUNCTION_NAME}


.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	echo "make would usually be run from the parent directory rather than here!\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: gomod test ## Build the gRPC service container locally after running unit tests

.PHONY: deploy
deploy: ## Deploy the the trigger Cloud Function
	gcloud functions deploy $(FUNCTION_NAME) --gen2 --region $(GCP_REGION) --runtime $(RUNTIME) \
     --entry-point=$(ENTRY_POINT) --trigger-http --allow-unauthenticated --ingress-settings=internal-only
	-gcloud pubsub subscriptions create ${SUBSCRIPTION_ID} --topic-project=${PROJECT_ID} --topic=${SHOPPING_CART_TOPIC} \
		--push-endpoint=$(gcloud functions describe ${FUNCTION_NAME} --gen2 --region=${GCP_REGION} --format="value(serviceConfig.uri)")
	# Expect an error in the line above - it will always fail if the subscription already exists
	# TODO: Implement authentication for the OrderFromCart Cloud Task target function


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
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/order
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/pb
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/types
	go mod edit -droprequire=github.com/mikebway/poc-gcp-ecomm/util
	go mod tidy
