all:
	go fmt
	go vet
	go clean
	godep go test -v
	godep go build
