all:
	go clean
	godep go test
	godep go build
