ALARM_MONITORING_NAME 			:= alarm_monitoring
FUNC_TEST_NAME         		    := test
NAMESPACE	   									:= "default"
CONFIG         								:= $(wildcard local.yml)
PKG            								:= github.com/eugeneverywhere/canopsis-test
PKG_LIST       								:= $(shell go list ${PKG}/... | grep -v /vendor/)

all: setup test build

setup: ## Installing all service dependencies.
	echo "Setup..."
	GO111MODULE=on go mod vendor

configure: ## Creating the local config yml.
	echo "Creating local config yml ..."
	cp config.example.yml local.yml

build: ## Build the executable file of service.
	echo "Building..."
	cd cmd/$(ALARM_MONITORING_NAME) && go build

run: ## Run service with local config.
	make build
	echo "Running..."
	cd cmd/$(ALARM_MONITORING_NAME) && ./$(ALARM_MONITORING_NAME) -config=../../local.yml

ft\:build: ## Build the executable file of service.
	echo "Building..."
	cd cmd/$(FUNC_TEST_NAME) && go build
ft\:run: ## Run service with local config.
	make build
	echo "Running..."
	cd cmd/$(FUNC_TEST_NAME) && ./$(FUNC_TEST_NAME) -config=../../local.yml

test: ## Run tests for all packages.
	echo "Testing..."
	go test -race ${PKG_LIST}

clean: ## Cleans the temp files and etc.
	echo "Clean..."
	rm -f cmd/$(ALARM_MONITORING_NAME)/$(ALARM_MONITORING_NAME)

help: ## Display this help screen
	grep -E '^[a-zA-Z_\-\:]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ": .*?## "}; {gsub(/[\\]*/,""); printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
