.PHONY: all
all: lint serve

.PHONY: lint
lint:
	GOGC=off golangci-lint run
	cd websocket-server && npm run lint
	cd editor && npm run lint

.PHONY: format
format:
	go fmt ./...
	cd websocket-server && npm run format
	cd editor && npm run format

.PHONY: serve
serve:
	CGO_ENABLED=0 modd

.PHONY: build
build:
	go build -v -o lakehousewiki ./cmd/server/main.go
	cd editor && npm run build

.PHONY: test
test:
	go test -v ./...
