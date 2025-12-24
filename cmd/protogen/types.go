package main

// TypeInfo contains parsed information about a struct type.
type TypeInfo struct {
	Name   string
	Fields []*FieldInfo
}

// FieldInfo contains parsed information about a struct field.
type FieldInfo struct {
	Name          string
	GoType        string
	FieldNum      int
	ProtoType     string
	IsRepeated    bool
	IsMessage     bool
	IsPointer     bool   // Field is a pointer type (*Type)
	IsSliceOfPtr  bool   // Field is a slice of pointers ([]*Type)
	IsOptional    bool   // Field is optional (can be nil/unset)
	IsEnum        bool   // Field is an enum type
	IsMap         bool   // Field is a map type
	IsCustom      bool   // Field uses custom marshaler interface (external types)
	ElemType      string // For slices, the element type (without [] or *)
	RawElemType   string // For slices, the raw element type (with * if applicable)
	BaseType      string // The base type without * or []
	NeedsTypeConv bool   // Needs type conversion (e.g., enum)
	ConvType      string // Type to convert to/from (e.g., int32 for enum)

	// Map-specific fields
	MapKeyType     string // Go type of map key (e.g., "string", "int32")
	MapValueType   string // Go type of map value (e.g., "int32", "*Sample")
	MapKeyProto    string // Proto type of map key (e.g., "string", "int32")
	MapValueProto  string // Proto type of map value (e.g., "int32", "message")
	MapValueIsMsg  bool   // Map value is a message type
	MapValueIsPtr  bool   // Map value is a pointer to message
	MapValueCustom bool   // Map value uses custom marshaler interface

	// Oneof-specific fields (for interface fields with multiple concrete types)
	IsOneof       bool           // Field is a oneof (interface with known implementations)
	OneofVariants []OneofVariant // List of concrete types and their field numbers
}

// OneofVariant represents a concrete type that can be stored in a oneof field
type OneofVariant struct {
	TypeName string // The concrete type name (e.g., "TextMessage")
	FieldNum int    // The protobuf field number for this variant
}
