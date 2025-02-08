# Build and run

## Bootstrap

### install protoc compiler
Download from https://github.com/protocolbuffers/protobuf/releases and install to PATH

### install go compiler for protobuf and gRPC

``` bash
sudo go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
sudo go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Build (manually)

1. compile .proto files: `protoc --go_out=internal/pb --go_opt=paths=source_relative --go-grpc_out=internal/pb --go-grpc_opt=paths=source_relative api/proto/file-walker-service.proto`
2. get dependencies: `go get`
3. build server:

``` bash
cd cmd/server
# build for macos
go --GOOS=darwin --GOARCH=arm64 build
```

4. build client

``` bash
cd cmd/client
go build
```
