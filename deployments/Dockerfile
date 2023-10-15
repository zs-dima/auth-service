FROM golang:1.21 AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the local package files to the container's workspace.
COPY . .

# Generate gRPC code
RUN apt-get update && apt-get install -y protobuf-compiler
RUN go get google.golang.org/grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
COPY proto proto
RUN protoc -Iproto proto/**.proto \
    --go_out=paths=source_relative:./internal/gen/proto \
    --go-grpc_out=paths=source_relative:./internal/gen/proto \
    --grpc-gateway_out ./internal/gen/proto --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true

# Generate SQL code
RUN apt-get update && apt-get install -y wget unzip && \
    wget https://github.com/sqlc-dev/sqlc/releases/download/v1.22.0/sqlc_1.22.0_linux_amd64.zip && \
    unzip sqlc_1.22.0_linux_amd64.zip && \
    mv sqlc /usr/local/bin/
COPY ./sqlc.yaml ./
COPY ./db/ ./db/
RUN sqlc generate

# Build the Go app for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o authservice ./cmd/main.go

FROM alpine:latest

# Certificates for gRPC communication
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the `builder` image to the current container
COPY --from=builder /app/authservice .

EXPOSE 8000 8080

CMD ["./authservice"]
