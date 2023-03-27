.PHONY: all
all: lint serve

.PHONY: lint
lint:
	GOGC=off golangci-lint run
	cd websocket-server && npm run lint
	cd websocket-client && npm run lint

.PHONY: format
format:
	go fmt ./...
	cd websocket-server && npm run format
	cd websocket-client && npm run format

.PHONY: serve
serve:
	CGO_ENABLED=0 modd

.PHONY: build
build:
	go build -v -o lakehouse ./cmd/server/main.go
	cd websocket-client && npm install && npm run build

.PHONY: test
test:
	go test -v ./...

.PHONY: deploy
deploy:
	GOOS=linux GOARCH=amd64 go build -v -o lakehouse ./cmd/server/main.go
	scp ./lakehouse deploy@5.75.194.9:/var/www/lakehouse/
	ssh deploy@5.75.194.9 sudo systemctl restart lakehouse-web
