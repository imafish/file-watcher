# Build and run

## Bootstrap

### install protoc compiler

Download from [protoc_compiler_github](https://github.com/protocolbuffers/protobuf/releases) and install to PATH

### install go compiler for protobuf and gRPC

``` bash
sudo go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
sudo go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Build (manually)

1. compile .proto files:

``` bash
protoc --go_out=. --go-grpc_out=. api/proto/file-watcher-service.proto
```

2. build server:

``` bash
cd cmd/server
# build for macos
go --GOOS=darwin --GOARCH=arm64 build
```

3. build client

``` bash
cd cmd/client
go build
```
