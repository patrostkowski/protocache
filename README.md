
# Protocache

**Protocache** is a lightweight in-memory cache server with gRPC support.

## 🏁 Getting Started

### 1. Install Dependencies

- Go (1.20+)
- `protoc` (Protocol Buffers compiler)
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins

### 2. Generate Protobuf Files

```bash
protoc --go_out=. --go-grpc_out=. api/cache.proto
```

Ensure your `cache.proto` has a correct `go_package` directive like:

```proto
option go_package = "github.com/patrostkowski/protocache/api/pb;cache";
```

### 3. Run the Server

```bash
go run ./cmd/protocache
```

---

## 🔌 Using grpcurl

### 📦 Server Reflection

To use `grpcurl`, make sure to enable server reflection in `main.go`:

```go
import "google.golang.org/grpc/reflection"

reflection.Register(grpcServer)
```

### 🔍 List Services

```bash
grpcurl -plaintext localhost:8080 list
```

### 📋 List Methods

```bash
grpcurl -plaintext localhost:8080 list cache.CacheService
```

### 🧪 Call RPCs

#### Set

```bash
grpcurl -plaintext -d '{"key":"foo","value":"YmFy"}' localhost:8080 cache.CacheService/Set
```

> Note: `value` must be base64-encoded (`"bar"` = `"YmFy"`)

#### Get

```bash
grpcurl -plaintext -d '{"key":"foo"}' localhost:8080 cache.CacheService/Get
```

#### Delete

```bash
grpcurl -plaintext -d '{"key":"foo"}' localhost:8080 cache.CacheService/Delete
```

#### Clear

```bash
grpcurl -plaintext -d '{}' localhost:8080 cache.CacheService/Clear
```

### 🧵 Use Stdin Input

```bash
echo '{"key":"foo"}' | grpcurl -plaintext -d @ localhost:8080 cache.CacheService/Get
```

---

## 📁 Project Structure

```
protocache/
├── api/
│   ├── cache.proto
│   └── pb/
├── cmd/
│   └── protocache/
│       └── main.go
├── internal/
│   ├── grpcserver/
│   └── cache/
├── go.mod
└── README.md
```
