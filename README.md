# easyproto-gen

Code generator for [easyproto](https://github.com/VictoriaMetrics/easyproto) marshal/unmarshal methods.

## When to Use This

Use this when you want **protobuf as a binary encoding format** (like JSON, but smaller and faster) and:

- You **control both ends** (client and server are both your code)
- You **don't need `.proto` files** or schema sharing with external systems
- You want to **define types in Go** and derive the wire format from them

This is the protobuf equivalent of using `encoding/json` with struct tags - you define Go structs, add tags, and get serialization for free.

### Don't Use This When

- You need to share `.proto` schemas with external teams or systems
- You're implementing a standard protocol (gRPC, OpenTelemetry, Prometheus remote write)
- You need strict schema evolution guarantees across multiple codebases

For those cases, use standard protoc tooling or write easyproto code manually.

## Benchmarks

Compared against `google.golang.org/protobuf` and `encoding/json` (Apple M2 Pro):

### Speed (ns/op, lower is better)

| Operation | easyproto-gen | google/protobuf | encoding/json |
|-----------|---------------|-----------------|---------------|
| Marshal   | **98 ns**     | 199 ns          | 367 ns        |
| Unmarshal | **86 ns**     | 459 ns          | 2067 ns       |
| Roundtrip | **174 ns**    | 635 ns          | 2488 ns       |

### Memory (allocs/op, lower is better)

| Operation | easyproto-gen | google/protobuf | encoding/json |
|-----------|---------------|-----------------|---------------|
| Marshal   | **0 allocs**  | 1 alloc         | 1 alloc       |
| Unmarshal | **1 alloc**   | 11 allocs       | 17 allocs     |
| Roundtrip | **1 alloc**   | 12 allocs       | 18 allocs     |

### Encoded Size

| Format | Size | vs JSON |
|--------|------|---------|
| protobuf (both) | 124 bytes | 61% |
| JSON | 202 bytes | 100% |

**Summary**: easyproto-gen is ~2x faster than google/protobuf for marshaling, ~5x faster for unmarshaling, and ~14x faster than JSON for roundtrips - with zero allocations on marshal.

Run benchmarks yourself: `go test ./bench -bench=. -benchmem`

## Quick Start

### 1. Install

```bash
go install github.com/aryehlev/easyproto-gen/cmd/protogen@latest
```

### 2. Add easyproto dependency

The generated code imports easyproto, so add it to your project:

```bash
go get github.com/VictoriaMetrics/easyproto
```

### 3. Define structs with tags

```go
//go:generate protogen -type=Message,User

type Message struct {
    ID        int64  `protobuf:"1"`
    Text      string `protobuf:"2"`
    Sender    *User  `protobuf:"3"`
    Timestamp int64  `protobuf:"4"`
}

type User struct {
    ID   int64  `protobuf:"1"`
    Name string `protobuf:"2"`
}
```

### 4. Generate

```bash
go generate ./...
```

### 5. Use

```go
msg := &Message{
    ID:   1,
    Text: "Hello",
    Sender: &User{ID: 42, Name: "Alice"},
    Timestamp: time.Now().Unix(),
}

// Encode (like json.Marshal)
data := msg.MarshalProtobuf(nil)

// Decode (like json.Unmarshal)
var msg2 Message
if err := msg2.UnmarshalProtobuf(data); err != nil {
    log.Fatal(err)
}
```

## Tag Format

```
`protobuf:"fieldNum[,type][,options]"`
```

**Basic usage** (types inferred from Go):
```go
Name  string  `protobuf:"1"`  // string
Count int32   `protobuf:"2"`  // int32
Data  []byte  `protobuf:"3"`  // bytes
Items []Item  `protobuf:"4"`  // repeated message
```

**Explicit wire types** (when encoding matters):
```go
Delta int32  `protobuf:"1,sint32"`  // zigzag for negative values
ID    uint64 `protobuf:"2,fixed64"` // fixed-width
```

**Options**:
- `enum` - enum type (int32 wire format)

## Type Mapping

| Go Type | Wire Type |
|---------|-----------|
| `string` | string |
| `[]byte` | bytes |
| `bool` | bool |
| `int32` | int32 |
| `int64`, `int` | int64 |
| `uint32` | uint32 |
| `uint64` | uint64 |
| `float32` | float |
| `float64` | double |
| `*T` | optional T |
| `[]T` | repeated T |
| `map[K]V` | map<K,V> |
| `struct` | message |

## Advanced

### Maps

```go
type Config struct {
    Labels map[string]string  `protobuf:"1"`
    Values map[int64]float64  `protobuf:"2"`
}
```

### Enums

```go
type Status int32

type Order struct {
    Status Status `protobuf:"1,enum"`
}
```

### Oneof (polymorphic fields)

```go
type Payload interface{ PayloadType() string }

type TextPayload struct {
    Text string `protobuf:"1"`
}
func (TextPayload) PayloadType() string { return "text" }

type BinaryPayload struct {
    Data []byte `protobuf:"1"`
}
func (BinaryPayload) PayloadType() string { return "binary" }

type Message struct {
    Content Payload `protobuf:"oneof,TextPayload:1,BinaryPayload:2"`
}
```

## CLI

```
protogen -type=Type1,Type2 [-output=file.go] [-noheader]

Flags:
  -type      Comma-separated struct names (required)
  -output    Output file (default: <type>_proto.go or <pkg>_proto.go)
  -noheader  Skip pool/interface declarations (for multiple generate calls)
```
