#!/bin/bash

# Go program does not really need to be compiled; use "go run" will be fine.
cd main
go build -mod=vendor ./main.go

