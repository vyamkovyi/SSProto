#!/bin/bash
go build --ldflags="-s -w" --buildmode=pie
scp -P 42 ./ssserver hexawolf@doggy:~/server/ssserver
