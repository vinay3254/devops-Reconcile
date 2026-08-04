#!/usr/bin/env bash
set -e

export GOROOT="$HOME/tools/go"
export GOPATH="$HOME/go"
export PATH="$HOME/tools/go/bin:$HOME/tools/bin:$HOME/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

cd /mnt/c/Users/Admin/Desktop/devops/api

echo "=== Tidy Go Modules ==="
go mod tidy

echo "=== Build API Service Binary ==="
mkdir -p bin
go build -o bin/api ./cmd/main.go

echo "API_BUILD_SUCCESS"
