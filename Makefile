.PHONY: install-grpc proto run

IDS := "0" "1" "2" "3" "4" "5" "6" "7" "8" "9"

install-grpc:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto/cs.pb.go: proto/cs.proto
	protoc --go_out=. --go_opt=paths=source_relative proto/cs.proto
proto/cs_grpc.pb.go: proto/cs.proto
	protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/cs.proto

proto: proto/cs.pb.go proto/cs_grpc.pb.go

run:
	@echo running 10 servers $(foreach id,$(IDS),\
		& (CS_ID=$(id) go run server/*.go) \
	);