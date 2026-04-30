#!/bin/bash

if ! command -v modview &> /dev/null
then
    echo "modview CLI could not be found. Installing..."
    go install github.com/bayraktugrul/modview@latest
fi

modview ./...
