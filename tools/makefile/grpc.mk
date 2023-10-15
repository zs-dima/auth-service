.PHONY: proto proto-lib

proto:
	@echo "Running proto codegeneration"
	@go get google.golang.org/grpc@latest	
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@protoc -Iproto proto/**.proto \
		--go_out=paths=source_relative:./internal/gen/proto \
		--go-grpc_out=paths=source_relative:./internal/gen/proto \
		--grpc-gateway_out ./internal/gen/proto --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true
