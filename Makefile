release: tidy lint doc clean build-windows-amd64 build-linux-amd64 build-linux-arm build-darwin-amd64 build-darwin-arm64

build-windows-amd64:
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
	zip bin/amqp-message-remover-windows-amd64.zip amqp-message-remover.exe
	go clean

build-linux-amd64:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	zip bin/amqp-message-remover-linux-amd64.zip amqp-message-remover
	go clean

build-linux-arm:
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build
	zip bin/amqp-message-remover-linux-arm.zip amqp-message-remover
	go clean

build-darwin-arm64:
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build
	zip bin/amqp-message-remover-darwin-arm64.zip amqp-message-remover
	go clean

build-darwin-amd64:
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
	zip bin/amqp-message-remover-darwin-amd64.zip amqp-message-remover
	go clean

lint:
	golangci-lint run ./... --fix

doc: build
	./amqp-message-remover doc doc

build: vendor
	go build

vendor: tidy
	go mod vendor

tidy:
	go mod tidy

mod-verify:
	go mod verify

clean:
	go clean