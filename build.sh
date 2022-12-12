#!/usr/bin/env bash

set -eu

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

export BIN_PATH=_output/${GOOS}/${GOARCH}

mkdir -p ${BIN_PATH}

echo "Building ovn-kubernetes plugins ..."
CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -buildmode=plugin -ldflags "-s -w" -o ${BIN_PATH}/ovn_kubernetes_plugin.so main.go
