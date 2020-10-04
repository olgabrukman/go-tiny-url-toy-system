all: build test lint

build:
	go build -o bin/main main.go

clean:
	rm -f bin/*

test:
	go test ./... -v

ci-test: build
	go test -race $$(go list ./...) -v -coverprofile .testCoverage.txt

lint:
	$(GOPATH)/bin/golangci-lint run ./... --fast --enable-all
