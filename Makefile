DOCKER_IMAGE_NAME=patrostkowski/protocache

generate:
	protoc --go_out=. --go-grpc_out=. api/cache.proto

docker-build:
	docker build -t ${DOCKER_IMAGE_NAME} .

docker-run:
	docker run --rm --name protocache -p 8080:8080 ${DOCKER_IMAGE_NAME}