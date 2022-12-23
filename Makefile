# Function to recursively match a list of wildcard patterns over all subdirectories of the target,
rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

# List all of the protobuf schema files
PROTO_SOURCES := $(call rwildcard,api/mikebway,*.proto)

# List the generated gRPC service definition Go source files
PROTO_SERVICE_FILES := $(call rwildcard,api/mikebway,*_api.proto)
GO_SERVICE_INTERMEDIATE_1 := $(patsubst %_api.proto,%_api_grpc.pb.go,$(PROTO_SERVICE_FILES))
GO_SERVICE_FILES := $(patsubst api/mikebway/%,pb/%,$(GO_SERVICE_INTERMEDIATE_1))

# Define the default recipe if non is specified on the command line
.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	$(info NOTE The 'build' step invokes gcloud build and requires that the full project source be first committed and pushed to GitHub)
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: protobuf
protobuf: ${GO_SERVICE_FILES} ## Compile the protocol buffer / gRPC schema files

.PHONY: test
test: protobuf ## Compile code and run unit tests locally on all components that support them
	$(info running test)
	$(MAKE) -C cart test
	$(MAKE) -C carttrigger test
	$(MAKE) -C fulfillment test
	$(MAKE) -C order test
	$(MAKE) -C ordertrigger test
	$(MAKE) -C orderfromcart test
	$(MAKE) -C tasktrigger test
	$(MAKE) -C types test

.PHONY: build
build: protobuf ## Build all the project components (invoking gcloud build)
	$(info running build)
	$(MAKE) -C cart build
	$(MAKE) -C carttrigger build
	$(MAKE) -C fulfillment build
	$(MAKE) -C order build
	$(MAKE) -C ordertrigger build
	$(MAKE) -C orderfromcart build
	$(MAKE) -C tasktrigger build

.PHONY: deploy
deploy: ## Deploy all project components to Google Cloud (you MUST build first)
	$(info running deploy)
	$(MAKE) -C cart deploy
	$(MAKE) -C carttrigger deploy
	$(MAKE) -C fulfillment deploy
	$(MAKE) -C order deploy
	$(MAKE) -C ordertrigger deploy
	$(MAKE) -C orderfromcart deploy
	$(MAKE) -C tasktrigger deploy

.PHONY: firestore
firestore: ## Run the Firestore emulator
	gcloud emulators firestore start --host-port=[::1]:8219 --project=poc-gcp-ecomm

.PHONY: pubsub
pubsub: ## Run the Firestore emulator
	gcloud beta emulators pubsub start --host-port=[::1]:8085 --project=poc-gcp-ecomm

.PHONY: stop-firestore
stop-firestore: ## Shutdown the Firestore emulator if it is running
	ps aux | grep firestore | grep java | sort | head -n 1 | awk '{print $$2}' | xargs kill -9

.PHONY: stop-pubsub
stop-pubsub: ## Shutdown the Firestore emulator if it is running
	ps aux | grep pubsub | grep java | sort | head -n 1 | awk '{print $$2}' | xargs kill -9

.PHONY: stop-all
stop-all: stop-firestore stop-pubsub ## Shutdown all gcloud emulators

${GO_SERVICE_FILES}: $(PROTO_SOURCES) ## Generate the Cart gRPC Go source code
	$(info generating protocol buffer and gRPC service code)
	protoc --proto_path api --go_opt=module="github.com/mikebway/poc-gcp-ecomm/pb" --go_out="pb" --go-grpc_opt=module="github.com/mikebway/poc-gcp-ecomm/pb" --go-grpc_out="pb" $(PROTO_SOURCES)
