# Protocache

**Protocache** is a lightweight in-memory cache server with gRPC support.

## 🏁 Getting Started

### 1. Install Dependencies

- `Go` (1.24+)
- `protoc` (3.19+)
- `GNU Make` (4.0+)
- `Docker` (28.0+)

### 2. Build the binaries

```bash
make build-all
```

### 3. Run the Server

```bash
make run
```

---

## 🖥️ Using protocachecli

`protocachecli` is a command-line client for interacting with the cache server.

### ✅ Commands

```bash
protocachecli -host localhost -port 50051 set foo bar
protocachecli -host localhost -port 50051 get foo
protocachecli -host localhost -port 50051 del foo
protocachecli -host localhost -port 50051 clear
```

If the `value` contains binary or non-UTF-8 data, it will be shown in base64 format.

### ℹ️ Help

```bash
protocachecli --help
```

---

## 🔌 Using grpcurl

### 🔍 List Services

```bash
grpcurl -plaintext localhost:50051 list
```

### 📋 List Methods

```bash
grpcurl -plaintext localhost:50051 list cache.CacheService
```

---

## 🧪 Testing

### ✅ Unit and E2E Tests

```bash
make test         # Run all tests
make test-e2e     # Run E2E tests
make bench        # Run benchmarks
```

---

## 🐳 Using Docker

You can build and run Protocache via Docker.

### 🔧 Build the Docker Image

```bash
make docker-build
```

This uses the image name: `patrostkowski/protocache`

### 🚀 Run the Server in a Container

```bash
make docker-run
```

This runs the server and exposes port `50051` on your local machine.

## License

Protocache is licensed under the [Apache License 2.0](./LICENSE).
