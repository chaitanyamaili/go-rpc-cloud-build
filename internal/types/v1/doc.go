// Package v1 holds the type and service definitions
package v1

//go:generate protoc -I=${PROTOBUF_INCLUDES}:${PARENT_LOCATION}/api/protobuf --go_opt=paths=source_relative --go_out=. types.proto
//go:generate protoc -I=${PROTOBUF_INCLUDES}:${PARENT_LOCATION}/api/protobuf --go-grpc_opt=paths=source_relative --go-grpc_out=. service.proto
