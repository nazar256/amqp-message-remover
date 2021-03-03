all: doc

doc: build
	bin/remover doc doc

build: vendor
	go build -o bin/remover -trimpath
	golangci-lint run

vendor: tidy
	go mod vendor

tidy:
	go mod tidy