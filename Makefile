maDEP := $(shell command -v dep 2> /dev/null)
SUM := $(shell which shasum)

COMMIT := $(shell git rev-parse HEAD)
CAT := $(if $(filter $(OS),Windows_NT),type,cat)

bucket=github.com

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
NAME=eon
SERVER=eond
CLIENT=eoncli
CosmosSDK=v0.39.1
TENDERMINT=v0.33.8
# Iavl=v0.12.4

# the height of the 1st block is GenesisHeight+1
GenesisHeight=0

# process linker flags
ifeq ($(VERSION),)
    VERSION = $(COMMIT)
endif

build_tags = netgo

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))


ldflags = -X $(bucket)/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X $(bucket)/cosmos/cosmos-sdk/version.Name=$(NAME) \
  -X $(bucket)/cosmos/cosmos-sdk/version.ServerName=$(SERVER) \
  -X $(bucket)/cosmos/cosmos-sdk/version.ClientName=$(CLIENT) \
  -X $(bucket)/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
  -X $(bucket)/tendermint/tendermint/types.startBlockHeightStr=$(GenesisHeight) \


ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -ldflags '$(ldflags)'  -gcflags "all=-N -l"
BUILD_TESTNET_FLAGS := $(BUILD_FLAGS)

all: install

install: eon

eon:
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/eond
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/eoncli

testnet:
	go install -v $(BUILD_TESTNET_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/eond
	go install -v $(BUILD_TESTNET_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/eoncli

# test-unit:
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./app/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/backend/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/common/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/dex/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/distribution/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/genutil/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/gov/...
# #	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/order/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/params/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/staking/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/token/...
# 	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/upgrade/...
# #	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./vendor/github.com/cosmos/cosmos-sdk/x/mint/...

get_vendor_deps:
	@echo "--> Generating vendor directory via dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -vendor-only

update_vendor_deps:
	@echo "--> Running dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -update

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
.PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

cli:
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/eoncli

server:
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/eond

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s

build:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/eond.exe ./cmd/eond
	go build $(BUILD_FLAGS) -o build/eoncli.exe ./cmd/eoncli
else
	go build $(BUILD_FLAGS) -o build/eond ./cmd/eond
	go build $(BUILD_FLAGS) -o build/eoncli ./cmd/eoncli
endif

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-docker-onlifenode:
	$(MAKE) -C ./networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/eond/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/eond:Z eon/node testnet --v 5 -o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down


.PHONY: build
