build:
	@go build -o bin/blocker
	
test:
	@go test -v ./...

run: build
	@./bin/blocker

proto:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/*.proto	
	
.PHONY: proto