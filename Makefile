all:
	go clean
	godep go test -v
	godep go build
