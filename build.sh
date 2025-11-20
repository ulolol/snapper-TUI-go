#!/bin/bash

# Create build directory if it doesn't exist
mkdir -p build

echo "Building for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o build/snapper-tui-go-amd64 .
if [ $? -eq 0 ]; then
    echo "Successfully built build/snapper-tui-go-amd64"
else
    echo "Failed to build for amd64"
    exit 1
fi

echo "Building for linux/arm64..."
GOOS=linux GOARCH=arm64 go build -o build/snapper-tui-go-arm64 .
if [ $? -eq 0 ]; then
    echo "Successfully built build/snapper-tui-go-arm64"
else
    echo "Failed to build for arm64"
    exit 1
fi

echo "Builds completed successfully."
