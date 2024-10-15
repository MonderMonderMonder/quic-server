#!/bin/bash

# Version
git rev-parse HEAD > VERSION

wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version

cd server_impl
go build -o server main.go
mv server .. && cd .. && cd client_impl
go build -o client main.go
mv client .. && cd ..

# Add all necessary files to the zip
# Do NOT remove the three scripts
zip -r artifact.zip VERSION setup-env.sh run-client.sh run-server.sh server client
