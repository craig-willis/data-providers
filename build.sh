#!/bin/bash

BUILD_DATE=`date +%Y-%m-%d\ %H:%M`
VERSIONFILE="pkg/version/version.go"
VERSION="1.0.9"


if [ "$1" == "local" ] || [ "$1" == "docker" ]; then

    if [ -e "$VERSIONFILE" ]; then 
        rm $VERSIONFILE
    fi
    mkdir -p pkg/version
    echo "package version" > $VERSIONFILE
    echo "const (" >> $VERSIONFILE
    echo "  VERSION = \"$VERSION \"" >> $VERSIONFILE
    echo "  BUILD_DATE = \"$BUILD_DATE\"" >> $VERSIONFILE
    echo ")" >> $VERSIONFILE
    
    glide install --strip-vendor --strip-vcs --update-vendored
    
	if [ "$1" == "local" ]; then 
        UNAME=$(uname)
        if [ "$UNAME" == "Darwin" ]; then
	        OS="darwin"
        elif [ "$UNAME" == "Linux" ]; then
	        OS="linux"
        fi
        
        echo Building apiserver-$OS-amd64
        GOOS=$OS GOARCH=amd64 go build -o build/bin/data-provider-$OS-amd64 ./cmd/
        
	elif [ "$1" == "docker" ]; then 	

        echo Building apiserver-linux-amd64
        GOOS=linux GOARCH=amd64 go build -o build/bin/data-provider-linux-amd64 ./cmd/
        GOOS=darwin GOARCH=amd64 go build -o build/bin/data-provider-darwin-amd64 ./cmd/
    fi
    
    rm -r pkg/version
elif [ "$1" == "clean" ]; then
	rm -r build
	rm -r vendor/github.com vendor/golang.org vendor/gopkg.in vendor/k8s.io
fi

