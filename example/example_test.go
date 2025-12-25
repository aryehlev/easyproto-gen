package example_test

import (
	"fmt"
	"time"

	"github.com/aryehlev/easyproto-gen/example"
)

func Example() {
	// Create a message with nested user
	msg := &example.Message{
		ID:   1,
		Text: "Hello, World!",
		Sender: &example.User{
			ID:   42,
			Name: "Alice",
		},
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}

	// Marshal to protobuf bytes
	data := msg.MarshalProtobuf(nil)

	// Unmarshal back
	var decoded example.Message
	if err := decoded.UnmarshalProtobuf(data); err != nil {
		panic(err)
	}

	fmt.Printf("ID: %d\n", decoded.ID)
	fmt.Printf("Text: %s\n", decoded.Text)
	fmt.Printf("Sender: %s (ID: %d)\n", decoded.Sender.Name, decoded.Sender.ID)

	// Output:
	// ID: 1
	// Text: Hello, World!
	// Sender: Alice (ID: 42)
}

func ExampleMessage_MarshalProtobuf() {
	msg := &example.Message{
		ID:   100,
		Text: "Test message",
	}

	// MarshalProtobuf appends to the provided slice (or creates new if nil)
	data := msg.MarshalProtobuf(nil)
	fmt.Printf("Encoded %d bytes\n", len(data))

	// You can also append to existing data
	buf := make([]byte, 0, 256)
	buf = msg.MarshalProtobuf(buf)
	fmt.Printf("Encoded to pre-allocated buffer: %d bytes\n", len(buf))

	// Output:
	// Encoded 16 bytes
	// Encoded to pre-allocated buffer: 16 bytes
}

func ExampleMessage_UnmarshalProtobuf() {
	// First, create and marshal a message
	original := &example.Message{
		ID:   999,
		Text: "Roundtrip test",
		Sender: &example.User{
			ID:   1,
			Name: "Bob",
		},
	}
	data := original.MarshalProtobuf(nil)

	// Unmarshal into a new struct
	var decoded example.Message
	if err := decoded.UnmarshalProtobuf(data); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Message ID: %d\n", decoded.ID)
	fmt.Printf("Message Text: %s\n", decoded.Text)
	fmt.Printf("Has Sender: %v\n", decoded.Sender != nil)

	// Output:
	// Message ID: 999
	// Message Text: Roundtrip test
	// Has Sender: true
}