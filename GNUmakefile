HOSTNAME=registry.terraform.io
NAMESPACE=logicalclocks
NAME=hopsworksai
VERSION=0.2.1
BINARY=terraform-provider-${NAME}

default: build

generate: fmt
	go generate  ./...

build: generate
	@echo "Building source code ..."
	mkdir -p ./bin
	go build -o ./bin/${BINARY}

install: build lint test
	@echo "Installing provider for terraform 0.13+ into ~/.terraform.d/plugins ... "
	@mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/$(shell go version | awk '{print $$4}' | sed 's#/#_#')
	@mv ./bin/${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/$(shell go version | awk '{print $$4}' | sed 's#/#_#')

fmt:
	@echo "Formatting source code using gofmt"
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w
	terraform fmt -recursive examples
	terraform fmt -recursive test-fixtures

lint:
	@echo "Linting source code using golangci-ling"
	golangci-lint run ./...

test:
	go test ./... -v -race -coverprofile=coverage.txt -covermode=atomic $(TESTARGS) -parallel=4

coverage: test
	go tool cover -html coverage.txt 

testacc:
	./test-fixtures/run-acceptance-tests.sh

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test ./hopsworksai -v -sweep="all" $(SWEEPARGS) -timeout 60m

.PHONY: build install testacc generate test fmt lint sweep coverage
