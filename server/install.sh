#!/usr/bin/env sh

go build ./cmd/ssmachmos/ssmachmos.go
mv ./ssmachmos ${INSTALL_PATH:-"/usr/local/bin"}
