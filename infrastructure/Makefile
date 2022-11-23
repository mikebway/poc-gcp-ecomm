# Project Settings
PROJECT_ID := poc-gcp-ecomm
GCP_REGION := us-central1

# Function to recursively match a list of wildcard patterns over all subdirectories of the target,
rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

# List all of the protobuf schema files
PROTO_SOURCES := $(call rwildcard,../api/mikebway,*.proto)

# Protocol Buffer sources and destinations
SOURCE_PROTO_PATH := ../api
MERGED_PROTO_PATH := ${PWD}/protobuf
SOURCE_CART_PROTO := mikebway/cart/cart.proto
MERGED_CART_PROTO := ${MERGED_PROTO_PATH}/mikebway/cart/cart.pps

# Pub/Sub schema names
SHOPPING_CART_SCHEMA := shopping-cart

# Pub/Sub topic names
SHOPPING_CART_TOPIC := shopping-cart


.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	#
	# Running setup or teardown multiple times will not fail. If you add something
	# to setup, running it will report errors for the components without aborting.

.PHONY: setup
setup: ${MERGED_CART_PROTO} ## Setup the Google Cloud infrastructure
	# Convert the multi-file protobuf Shopping Cart schema to a single line and register as a Pub/Sub schema
	-TEMP=`tail -n +5 ${MERGED_CART_PROTO} | tr -d '[:cntrl:]' | tr '"' "'"`; gcloud pubsub schemas create ${SHOPPING_CART_SCHEMA} --quiet --type=PROTOCOL_BUFFER --definition="$$TEMP"

	# Declare the Pub/Sub topic that moves shopping carts to the Order system
	-gcloud pubsub topics create ${SHOPPING_CART_TOPIC} --quiet --schema ${SHOPPING_CART_SCHEMA} --message-retention-duration=7d --message-encoding=binary


.PHONY: teardown
teardown: ## Tear down the Google Cloud infrastructure
	-gcloud pubsub topics delete ${SHOPPING_CART_TOPIC} --quiet
	-gcloud pubsub schemas delete ${SHOPPING_CART_SCHEMA} --quiet

## Generate a single file protobuf schema for a shopping cart from the multi-file master definition
${MERGED_CART_PROTO}: ${PROTO_SOURCES}
	cd ${SOURCE_PROTO_PATH}; protoc ${SOURCE_CART_PROTO} --pubsub-schema_out=${MERGED_PROTO_PATH} --pubsub-schema_opt=schema-syntax=proto3