#!/bin/bash

cd "$(dirname "$0")"
cd ..

# Core Service
rm -rf ./app/service-core/proto
mkdir ./app/service-core/proto
protoc --go_out=./app/service-core/proto --go_opt=paths=source_relative \
    --go-grpc_out=./app/service-core/proto --go-grpc_opt=paths=source_relative \
    --proto_path=./proto \
    ./proto/*.proto
