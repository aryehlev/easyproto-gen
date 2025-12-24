package example

//go:generate go run ../cmd/protogen -type=Message,User

// Message represents a chat message.
type Message struct {
	ID        int64  `protobuf:"1"`
	Text      string `protobuf:"2"`
	Sender    *User  `protobuf:"3"`
	Timestamp int64  `protobuf:"4"`
}

// User represents a user.
type User struct {
	ID   int64  `protobuf:"1"`
	Name string `protobuf:"2"`
}
