#!/usr/bin/env bash

platform=$(uname -s)
arch=$(uname -m)

if [[ "$platform" == "Linux" ]]; then
  if [[ "$arch" == "aarch64" ]]; then
    go build -ldflags="-s -w" -o luffy.aarch64
    upx --best --lzma luffy.aarch64
  else if [[ "$arch" == "riscv64" ]]; then
    GOOS=linux GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o luffy.rv64
  else
    go build -ldflags="-s -w" -o luffy.amd64
    upx --best --lzma luffy.amd64
  fi
else
  go build -o luffy-macos.aarch64
fi

# GOOS=linux GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags="-s -w" -o luffy.rv64
