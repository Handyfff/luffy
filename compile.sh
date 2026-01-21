#!/usr/bin/env bash

platform=$(uname -s)

if [[ "$platform" == "Linux" ]]; then
  go build -ldflags="-s -w" -o luffy
  upx --best --lzma luffy
else
  go build -o luffy
fi
