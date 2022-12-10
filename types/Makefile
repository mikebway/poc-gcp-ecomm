.DEFAULT_GOAL := help

.PHONY: help
help: ## List of available commands
	echo "make would usually be run from the parent directory rather than here!\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: test
test: compile ## Run the unit tests locally
	go test ./... -coverprofile cover.out -race; \
   	go tool cover -func cover.out

.PHONY: compile
compile: ## Compile the Go code locally
	go build
