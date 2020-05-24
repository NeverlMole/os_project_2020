#!/bin/sh
protoc -I ./ ./log.proto --go_out=plugins=grpc:./go 
