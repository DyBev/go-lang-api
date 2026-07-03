CGO_ENABLED := 0
BUILD_DIR := ./build
GOOS := linux
GOARCH := arm64

run:
	CGO_ENABLED=$(CGO_ENABLED) go run .
