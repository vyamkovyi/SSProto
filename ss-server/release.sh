#!/bin/bash
go build --ldflags="-s -w"
scp -P 42 ./ssserver hexawolf@doggy:~/server/ssserver
