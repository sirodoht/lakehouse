.PHONY: all
all: lint serve

.PHONY: lint
lint:
	$(info Running Go linters)
	@GOGC=off golangci-lint run

.PHONY: format
format:
	go fmt ./...

.PHONY: serve
serve:
	CGO_ENABLED=0 modd
