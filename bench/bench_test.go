package bench

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"
)

// Test data
var (
	easyMsg = &Message{
		ID:        12345,
		Text:      "Hello, this is a test message with some content",
		Timestamp: 1703462400,
		Tags:      []string{"urgent", "notification", "system"},
		Sender: &User{
			ID:    42,
			Name:  "Alice Smith",
			Email: "alice@example.com",
		},
	}

	protoMsg = &ProtoMessage{
		Id:        12345,
		Text:      "Hello, this is a test message with some content",
		Timestamp: 1703462400,
		Tags:      []string{"urgent", "notification", "system"},
		Sender: &ProtoUser{
			Id:    42,
			Name:  "Alice Smith",
			Email: "alice@example.com",
		},
	}

	jsonMsg = easyMsg // same struct, uses json tags
)

// Pre-encoded data for unmarshal benchmarks
var (
	easyEncoded  []byte
	protoEncoded []byte
	jsonEncoded  []byte
)

func init() {
	easyEncoded = easyMsg.MarshalProtobuf(nil)
	protoEncoded, _ = proto.Marshal(protoMsg)
	jsonEncoded, _ = json.Marshal(jsonMsg)
}

// =============================================================================
// Marshal Benchmarks
// =============================================================================

func BenchmarkMarshal_Easyproto(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = easyMsg.MarshalProtobuf(buf[:0])
	}
}

func BenchmarkMarshal_GoogleProtobuf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(protoMsg)
	}
}

func BenchmarkMarshal_JSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(jsonMsg)
	}
}

// =============================================================================
// Unmarshal Benchmarks
// =============================================================================

func BenchmarkUnmarshal_Easyproto(b *testing.B) {
	b.ReportAllocs()
	var msg Message
	for i := 0; i < b.N; i++ {
		_ = msg.UnmarshalProtobuf(easyEncoded)
	}
}

func BenchmarkUnmarshal_GoogleProtobuf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var msg ProtoMessage
		_ = proto.Unmarshal(protoEncoded, &msg)
	}
}

func BenchmarkUnmarshal_JSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var msg Message
		_ = json.Unmarshal(jsonEncoded, &msg)
	}
}

// =============================================================================
// Size Comparison (run as test, not benchmark)
// =============================================================================

func TestEncodedSize(t *testing.T) {
	easySize := len(easyEncoded)
	protoSize := len(protoEncoded)
	jsonSize := len(jsonEncoded)

	t.Logf("Encoded sizes:")
	t.Logf("  easyproto:       %d bytes", easySize)
	t.Logf("  google/protobuf: %d bytes", protoSize)
	t.Logf("  encoding/json:   %d bytes", jsonSize)
	t.Logf("")
	t.Logf("Size comparison (vs JSON):")
	t.Logf("  easyproto:       %.1f%% of JSON size", float64(easySize)/float64(jsonSize)*100)
	t.Logf("  google/protobuf: %.1f%% of JSON size", float64(protoSize)/float64(jsonSize)*100)
}

// =============================================================================
// Roundtrip Benchmarks (marshal + unmarshal)
// =============================================================================

func BenchmarkRoundtrip_Easyproto(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	var msg Message
	for i := 0; i < b.N; i++ {
		buf = easyMsg.MarshalProtobuf(buf[:0])
		_ = msg.UnmarshalProtobuf(buf)
	}
}

func BenchmarkRoundtrip_GoogleProtobuf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data, _ := proto.Marshal(protoMsg)
		var msg ProtoMessage
		_ = proto.Unmarshal(data, &msg)
	}
}

func BenchmarkRoundtrip_JSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(jsonMsg)
		var msg Message
		_ = json.Unmarshal(data, &msg)
	}
}
