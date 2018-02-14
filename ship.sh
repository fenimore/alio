#!/bin/bash

echo "Moving alio to bin at $GOPATH/bin"
go build alio.go && mv alio $GOPATH/bin/
