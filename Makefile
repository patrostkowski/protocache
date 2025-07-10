DOCKER_IMAGE_NAME=patrostkowski/protocache

generate:
	protoc --go_out=. --go-grpc_out=. api/cache.proto

run:
	go run ./cmd/protocache

docker-build:
	docker build -t ${DOCKER_IMAGE_NAME} .

docker-run:
	docker run --rm --network="host"  --name protocache ${DOCKER_IMAGE_NAME}

test:
	go test ./...

test-e2e:
	go test ./tests/e2e

bench:
	go test -bench=. -benchmem ./tests/benchmark

create-cluster:
	kind create cluster