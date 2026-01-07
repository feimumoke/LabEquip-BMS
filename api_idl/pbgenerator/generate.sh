#!/bin/bash

# go get github.com/envoyproxy/protoc-gen-validate
# PATH=$PATH:$GOPATH/bin

for p in `find . -name "*.proto"`; do
  echo "$p"
  echo "protoc I .  --validate_out="lang=go,paths=source_relative:./"  --go_out=plugins=grpc,paths=source_relative:./ $p"
  protoc -I .  --validate_out="lang=go,paths=source_relative:./"  --go_out=plugins=grpc,paths=source_relative:./ $p
done
