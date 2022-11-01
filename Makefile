.PHONY: all
all: lint serve

.PHONY: lint
lint:
	$(info Running Go linters)
	@GOGC=off golangci-lint run

.PHONY: format
format:
	$(info Running go fmt)
	go fmt ./...

.PHONY: serve
serve:
	$(info Watch files and run server)
	CGO_ENABLED=0 modd

.PHONY: build
build:
	$(info Running go build)
	go build -v ./...

.PHONY: test
test:
	$(info Running go test)
	go test -v ./...
