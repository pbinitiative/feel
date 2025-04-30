GO_FILES := $(shell find . -name '*.go')
GO_FLAGS :=

all: build

test:
	go test $$(go list ./... | grep -v /tests/tck)

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
	rm -rf resources/
	rm -rf tests/tck/TestCases/

.PHONY: test gofmt build-cli clean
.SECONDARY: $(buildarchdirs)

.PHONY: extract-testcases
extract-testcases:
	mkdir -p resources/
	git clone --depth 1 https://github.com/dmn-tck/tck.git resources/tck/
	go run ./cmd/testcase-extractor --dir resources/tck --output-dir tests/tck/
	rm -rf resources/


.PHONY: test-tck
test-tck:
	go test ./tests/tck

