all: check

check:
	test -z $$(gofmt -d .)
	go vet ./...
