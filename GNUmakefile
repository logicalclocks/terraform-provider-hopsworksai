default: build

generate: fmt
	@echo "Generating documentations ..."
	go generate  ./...

build: generate lint
	@echo "Building source code ..."
	mkdir -p ./bin
	go build -o ./bin/terraform-provider-hopsworksai

fmt:
	@echo "Formatting source code using gofmt"
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w
	terraform fmt -recursive examples
	terraform fmt -recursive test-fixtures

lint:
	@echo "Linting source code using golangci-ling"
	golangci-lint run ./...

test:
	@echo "Running unit tests ..."
	go test ./... -v -race -coverprofile=coverage.txt -covermode=atomic $(TESTARGS) -parallel=4

coverage: test
	@echo "Running code coverage ..."
	go tool cover -html coverage.txt 

testacc:
	@echo "Running acceptance tests ..."
	./test-fixtures/run-acceptance-tests.sh

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test ./hopsworksai -v -sweep="all" $(SWEEPARGS) -timeout 60m

.PHONY: build testacc generate test fmt lint sweep coverage
