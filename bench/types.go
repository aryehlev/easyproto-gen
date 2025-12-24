package bench

//go:generate protogen -type=Message,User

// Message is the easyproto-gen version.
type Message struct {
	ID        int64    `protobuf:"1" json:"id"`
	Text      string   `protobuf:"2" json:"text"`
	Sender    *User    `protobuf:"3" json:"sender"`
	Timestamp int64    `protobuf:"4" json:"timestamp"`
	Tags      []string `protobuf:"5" json:"tags"`
}

// User is the easyproto-gen version.
type User struct {
	ID    int64  `protobuf:"1" json:"id"`
	Name  string `protobuf:"2" json:"name"`
	Email string `protobuf:"3" json:"email"`
}
