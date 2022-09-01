.PHONY: all fmt test stresstest tidy

all: fmt test tidy

fmt:
	go fmt ./...

test:
	go test -race -v -count=1 -cover .

stresstest:
	go test -race -failfast -count=1000 -cover .

tidy:
	go mod tidy