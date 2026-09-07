.PHONY: build clean test

BINARY := iam-migrate

build:
	go build -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	go test ./...
