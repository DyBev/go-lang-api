CGO_ENABLED := 0
BUILD_DIR := ./build
GOOS := linux
GOARCH := arm64

run:
	CGO_ENABLED=$(CGO_ENABLED) go run .

docker:
	docker image rm go-api --force
	docker build --tag go-api .
