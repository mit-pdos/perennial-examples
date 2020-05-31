all: check

check:
	@gofmt -d .
	test -z "$$(gofmt -d .)"
	go vet ./...
