#!/usr/bin/env sh

# protoc ./*.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

mkdir -p {addsvc,usersvc}

protoc --go_out=addsvc \
  --go_opt=paths=source_relative \
  --go-grpc_out=addsvc \
  --go-grpc_opt=paths=source_relative \
  ./addsvc.proto

protoc --go_out=usersvc \
  --go_opt=paths=source_relative \
  --go-grpc_out=usersvc \
  --go-grpc_opt=paths=source_relative \
  ./usersvc.proto