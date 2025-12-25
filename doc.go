// Package easyprotogen provides a code generator for [easyproto] marshal/unmarshal methods.
//
// This package generates high-performance protobuf serialization code for Go structs
// using struct tags, similar to how encoding/json works with json tags.
//
// # When to Use This
//
// Use this when you want protobuf as a binary encoding format and:
//   - You control both ends (client and server are both your code)
//   - You don't need .proto files or schema sharing with external systems
//   - You want to define types in Go and derive the wire format from them
//
// # Installation
//
// Install the code generator:
//
//	go install github.com/aryehlev/easyproto-gen/cmd/protogen@latest
//
// Add the easyproto runtime dependency to your project:
//
//	go get github.com/VictoriaMetrics/easyproto
//
// # Usage
//
// Define your structs with protobuf tags:
//
//	//go:generate protogen -type=Message,User
//
//	type Message struct {
//	    ID        int64  `protobuf:"1"`
//	    Text      string `protobuf:"2"`
//	    Sender    *User  `protobuf:"3"`
//	    Timestamp int64  `protobuf:"4"`
//	}
//
//	type User struct {
//	    ID   int64  `protobuf:"1"`
//	    Name string `protobuf:"2"`
//	}
//
// Then run go generate:
//
//	go generate ./...
//
// The generator creates MarshalProtobuf and UnmarshalProtobuf methods:
//
//	msg := &Message{ID: 1, Text: "Hello"}
//	data := msg.MarshalProtobuf(nil)
//
//	var msg2 Message
//	err := msg2.UnmarshalProtobuf(data)
//
// # Tag Format
//
// The protobuf tag format is:
//
//	`protobuf:"fieldNum[,type][,options]"`
//
// Field numbers are required. Types are inferred from Go types:
//
//	string    -> string       int32   -> int32      float32 -> float
//	[]byte    -> bytes        int64   -> int64      float64 -> double
//	bool      -> bool         uint32  -> uint32     CustomType -> message
//	int       -> int64        uint64  -> uint64     map[K]V -> map
//
// Explicit wire types can be specified when needed:
//
//	Delta int32  `protobuf:"1,sint32"`  // zigzag encoding for negative values
//	ID    uint64 `protobuf:"2,fixed64"` // fixed-width encoding
//
// Options include:
//   - enum: marks field as enum type (uses int32 wire format)
//
// # Performance
//
// Compared to google.golang.org/protobuf and encoding/json (Apple M2 Pro):
//
//	| Operation | easyproto-gen | google/protobuf | encoding/json |
//	|-----------|---------------|-----------------|---------------|
//	| Marshal   | 98 ns         | 199 ns          | 367 ns        |
//	| Unmarshal | 86 ns         | 459 ns          | 2067 ns       |
//
// easyproto-gen achieves zero allocations on marshal and minimal allocations on unmarshal.
//
// # CLI Reference
//
// The protogen command accepts the following flags:
//
//	protogen -type=Type1,Type2 [-output=file.go] [-noheader]
//
//	-type      Comma-separated struct names (required)
//	-output    Output file (default: <type>_proto.go or <pkg>_proto.go)
//	-noheader  Skip pool/interface declarations (for multiple generate calls)
//
// [easyproto]: https://github.com/VictoriaMetrics/easyproto
package easyprotogen
