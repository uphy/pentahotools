#!/bin/bash

mkdir -p dist
mkdir -p dist/static

# Build web component
which yarn > /dev/null 2>&1
if [ "$NOWEB" == "" -a $? == 0 ]; then
    pushd . > /dev/null
    cd web
    yarn build
    cp -rp dist/* ../dist/static
    popd
fi

# Build pentahotools
go build -o dist/pentahotools main.go

# Generate archive
cd dist
tar zcf ../pentahotools.tar.gz *

