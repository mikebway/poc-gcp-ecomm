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
MERGED_CART_PROTO := ${MERGED_PROTO_PATH}/${SOURCE_CART_PROTO}

.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	echo "make would usually be run from the parent directory rather than here!\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: setup
setup: protobuf ## Setup the Google Cloud infrastructure

.PHONY: teardown
teardown: ## Tear down the Google Cloud infrastructure

.PHONY: protobuf
protobuf: ${MERGED_CART_PROTO} ## Generate the single file protobuf schemas for all out Pub/Sub topics
	echo ${PROTO_SOURCES}

${MERGED_CART_PROTO}: ${PROTO_SOURCES} ## Generate the single file protobuf schema for a shopping cart
	cd ${SOURCE_PROTO_PATH}; protoc ${SOURCE_CART_PROTO} --pubsub-schema_out=${MERGED_PROTO_PATH} --pubsub-schema_opt=schema-syntax=proto3