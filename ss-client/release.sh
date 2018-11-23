#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Usage: ./release.sh CERTIFICATE SERVER-ADDRESS"
    echo "E.g. ./release.sh ./mc.pem hexawolf.me:48879"
    exit 1
fi

certPath="$1"
serverName="$2"

CGO_ENABLED=0 GOOS=linux ./build.sh $certPath $serverName "Updater-linux"
GOARCH=386 GOOS=windows ./build.sh $certPath $serverName "Updater.exe"
CGO_ENABLED=0 GOOS=darwin ./build.sh $certPath $serverName "Updater-mac"
