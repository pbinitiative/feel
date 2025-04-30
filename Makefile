RESOURCES_DIR=resources
TCK_DIR=$(RESOURCES_DIR)/tck
TCK_REPO=https://github.com/dmn-tck/tck.git
TEST_CONFIG_FILE=./tck-tests/test-cases_feel.yaml

GO_FILES := $(shell find . -name '*.go')
GO_FLAGS :=

all: build

test:
	go test -v ./...

cover:
	go test -coverprofile=coverage.out  ./...
	@echo To view coverage graph use go tool cover -html=coverage.out

golint:
	go fmt ./...
	go vet ./...

build: build-cli

build-cli: bin/feel

bin/feel: ${GO_FILES}
	go build $(GO_FLAGS) -o $@ cli/main.go

clean:
	rm -rf build dist bin/*
	rm -rf $(TCK_DIR)

.PHONY: test gofmt build-cli clean
.SECONDARY: $(buildarchdirs)

.PHONY: extract-testcases
extract-testcases:
	mkdir -p $(RESOURCES_DIR)
	git clone --depth 1 $(TCK_REPO) $(TCK_DIR)
	go run ./cmd/testcase-extractor --dir $(TCK_DIR) --output-file $(TEST_CONFIG_FILE)
	rm -rf $(TCK_DIR)

.PHONY: tck-tests
tck-tests:
	go test -tags tck_feel_test ./tck-tests

