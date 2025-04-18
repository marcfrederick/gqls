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
	# go-check-sumtype is included in golangci-lint, but doesn't work properly,
	# so we run it separately
	go tool go-check-sumtype $$(go list ./... | grep -v vendor)

test: generate
	go test -v ./...
