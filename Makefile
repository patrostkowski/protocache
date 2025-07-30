BUILD_BIN_DIR=./bin
PWD := $(shell pwd)
DOCKER_IMAGE_NAME=patrostkowski/protocache

.PHONY: all generate run build-all build build-cli docker-build docker-run test test-e2e bench create-cluster clean

all: build-all

generate:
	protoc \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		internal/api/cache/v1alpha/cache.proto

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
	docker run --rm --network="host" \
	--name protocache \
	-v "$(PWD)/example/config.yaml:/etc/protocache/config.yaml:ro,z" \
	${DOCKER_IMAGE_NAME}

lint-verify:
	golangci-lint config verify

lint:
	golangci-lint run

check: lint-verify lint test

license:
	addlicense -c "Patryk Rostkowski" -l apache -y 2025 .

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
