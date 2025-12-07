#!/bin/sh

GOOS=js GOARCH=wasm go build -o wasm/go-js.wasm ./cmd/go-js-wasm "$@"

cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" wasm/wasm_exec.js
