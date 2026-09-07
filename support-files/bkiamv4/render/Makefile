.PHONY: build clean test

BINARY := iam-render

build:
	go build -o $(BINARY) .

clean:
	rm -f $(BINARY)

test:
	cd .. && go test -v ./...
