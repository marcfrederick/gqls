BINARY := gqls
ENTRYPOINT := cmd/cli/main.go

.PHONY: all clean generate build lint test

all: generate build lint test

clean:
	rm -rf ./$(BINARY)

generate: internal/gql/grammar
	go generate ./...

build: generate
	go build -o $(BINARY) $(ENTRYPOINT)

lint:
	golangci-lint run ./...

test: generate
	go test -v ./...
