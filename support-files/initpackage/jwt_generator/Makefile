.PHONY: build clean test

BINARY := jwt-generator

build:
	go build -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	cd .. && go test -v ./...
