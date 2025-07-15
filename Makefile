BUILD_BIN_DIR=./bin

DOCKER_IMAGE_NAME=patrostkowski/protocache

.PHONY: all generate run build-all build build-cli docker-build docker-run test test-e2e bench create-cluster clean

all: build-all

generate:
	protoc --go_out=. --go-grpc_out=. api/cache.proto

run:
	$(MAKE) build
	exec ${BUILD_BIN_DIR}/protocache

build-all:
	$(MAKE) build
	$(MAKE) build-cli

build:
	go build -o ${BUILD_BIN_DIR}/protocache ./cmd/protocache

build-cli:
	go build -o ${BUILD_BIN_DIR}/protocachecli ./cmd/protocachecli

docker-build:
	docker build -t ${DOCKER_IMAGE_NAME} .

docker-run:
	docker run --rm -p 50051:50051 -p 9091:9091 --name protocache ${DOCKER_IMAGE_NAME} -id single

test:
	go test ./...

test-e2e:
	go test ./tests/e2e

bench:
	go test -bench=. -benchmem ./tests/benchmark

create-cluster:
	kind create cluster

clean:
	rm -rf ${BUILD_BIN_DIR}
	git clean -xdff
