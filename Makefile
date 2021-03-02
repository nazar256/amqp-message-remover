doc: build
	bin/remover doc doc

build: vendor
	go build -o bin/remover

vendor: tidy
	go mod vendor

tidy:
	go mod tidy