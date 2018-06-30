#!/bin/bash
go build --ldflags="-s -w" --buildmode=pie
GOOS=windows go build --ldflags="-s -w"
