.PHONY: all fmt test tidy

all: fmt test tidy

fmt:
	go fmt ./...

test:
	go test -race -v -count=1 -cover .

tidy:
	go mod tidy