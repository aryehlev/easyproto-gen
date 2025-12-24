package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// parseTestStruct parses a struct definition from source code and returns the TypeInfo
func parseTestStruct(t *testing.T, typeName, source string) (*TypeInfo, error) {
	t.Helper()
	fset := token.NewFileSet()
	src := "package test\n\n" + source
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != typeName {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("type %s is not a struct", typeName)
			}
			return parseStruct(typeName, structType)
		}
	}
	t.Fatalf("type %s not found", typeName)
	return nil, nil
}

func TestOneofTagParsing_ValidTag(t *testing.T) {
	source := `
type Message interface{ MessageType() string }
type TextMessage struct{}
type ImageMessage struct{}

type Chat struct {
	ID      int64   ` + "`protobuf:\"1\"`" + `
	Content Message ` + "`protobuf:\"oneof,TextMessage:2,ImageMessage:3\"`" + `
}
`
	info, err := parseTestStruct(t, "Chat", source)
	if err != nil {
		t.Fatalf("expected valid oneof tag, got error: %v", err)
	}

	if len(info.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(info.Fields))
	}

	// Find the Content field
	var contentField *FieldInfo
	for _, f := range info.Fields {
		if f.Name == "Content" {
			contentField = f
			break
		}
	}

	if contentField == nil {
		t.Fatal("Content field not found")
	}

	if !contentField.IsOneof {
		t.Error("expected IsOneof to be true")
	}

	if len(contentField.OneofVariants) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(contentField.OneofVariants))
	}

	if contentField.OneofVariants[0].TypeName != "TextMessage" || contentField.OneofVariants[0].FieldNum != 2 {
		t.Errorf("variant 0 mismatch: got %+v", contentField.OneofVariants[0])
	}
	if contentField.OneofVariants[1].TypeName != "ImageMessage" || contentField.OneofVariants[1].FieldNum != 3 {
		t.Errorf("variant 1 mismatch: got %+v", contentField.OneofVariants[1])
	}
}

func TestOneofTagParsing_MissingVariants(t *testing.T) {
	source := `
type Message interface{}
type Chat struct {
	Content Message ` + "`protobuf:\"oneof\"`" + `
}
`
	_, err := parseTestStruct(t, "Chat", source)
	if err == nil {
		t.Error("expected error for oneof with no variants")
	}
	if !strings.Contains(err.Error(), "at least one variant") {
		t.Errorf("expected 'at least one variant' error, got: %v", err)
	}
}

func TestOneofTagParsing_InvalidVariantFormat(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantErr string
	}{
		{
			name:    "missing colon",
			tag:     "`protobuf:\"oneof,TextMessage\"`",
			wantErr: "expected Type:FieldNum format",
		},
		{
			name:    "non-numeric field number",
			tag:     "`protobuf:\"oneof,TextMessage:abc\"`",
			wantErr: "invalid field number",
		},
		{
			name:    "field number too low",
			tag:     "`protobuf:\"oneof,TextMessage:0\"`",
			wantErr: "must be 1-536870911",
		},
		{
			name:    "field number too high",
			tag:     "`protobuf:\"oneof,TextMessage:536870912\"`",
			wantErr: "must be 1-536870911",
		},
		{
			name:    "reserved field number",
			tag:     "`protobuf:\"oneof,TextMessage:19000\"`",
			wantErr: "reserved",
		},
		{
			name:    "reserved field number high end",
			tag:     "`protobuf:\"oneof,TextMessage:19999\"`",
			wantErr: "reserved",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := `
type Message interface{}
type TextMessage struct{}
type Chat struct {
	Content Message ` + tc.tag + `
}
`
			_, err := parseTestStruct(t, "Chat", source)
			if err == nil {
				t.Error("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestOneofTagParsing_DuplicateFieldNumbers(t *testing.T) {
	source := `
type Message interface{}
type TextMessage struct{}
type ImageMessage struct{}
type Chat struct {
	Content Message ` + "`protobuf:\"oneof,TextMessage:2,ImageMessage:2\"`" + `
}
`
	_, err := parseTestStruct(t, "Chat", source)
	if err == nil {
		t.Error("expected error for duplicate field numbers in oneof")
	}
	if !strings.Contains(err.Error(), "duplicate field number") {
		t.Errorf("expected 'duplicate field number' error, got: %v", err)
	}
}

func TestOneofTagParsing_DuplicateWithRegularField(t *testing.T) {
	source := `
type Message interface{}
type TextMessage struct{}
type Chat struct {
	ID      int64   ` + "`protobuf:\"2\"`" + `
	Content Message ` + "`protobuf:\"oneof,TextMessage:2\"`" + `
}
`
	_, err := parseTestStruct(t, "Chat", source)
	if err == nil {
		t.Error("expected error for field number collision between oneof variant and regular field")
	}
	if !strings.Contains(err.Error(), "duplicate field number") {
		t.Errorf("expected 'duplicate field number' error, got: %v", err)
	}
}

func TestOneofTagParsing_MultipleVariants(t *testing.T) {
	source := `
type Message interface{}
type TextMessage struct{}
type ImageMessage struct{}
type VideoMessage struct{}
type AudioMessage struct{}
type Chat struct {
	Content Message ` + "`protobuf:\"oneof,TextMessage:1,ImageMessage:2,VideoMessage:3,AudioMessage:4\"`" + `
}
`
	info, err := parseTestStruct(t, "Chat", source)
	if err != nil {
		t.Fatalf("expected valid oneof tag with 4 variants, got error: %v", err)
	}

	if len(info.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(info.Fields))
	}

	contentField := info.Fields[0]
	if len(contentField.OneofVariants) != 4 {
		t.Fatalf("expected 4 variants, got %d", len(contentField.OneofVariants))
	}

	expected := []struct {
		name     string
		fieldNum int
	}{
		{"TextMessage", 1},
		{"ImageMessage", 2},
		{"VideoMessage", 3},
		{"AudioMessage", 4},
	}

	for i, exp := range expected {
		if contentField.OneofVariants[i].TypeName != exp.name {
			t.Errorf("variant %d name: got %q, want %q", i, contentField.OneofVariants[i].TypeName, exp.name)
		}
		if contentField.OneofVariants[i].FieldNum != exp.fieldNum {
			t.Errorf("variant %d field num: got %d, want %d", i, contentField.OneofVariants[i].FieldNum, exp.fieldNum)
		}
	}
}

func TestOneofTagParsing_ValidFieldNumberEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		fieldNum string
		want     int
	}{
		{"minimum field number", "1", 1},
		{"maximum field number", "536870911", 536870911},
		{"just below reserved", "18999", 18999},
		{"just above reserved", "20000", 20000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := `
type Message interface{}
type TextMessage struct{}
type Chat struct {
	Content Message ` + "`protobuf:\"oneof,TextMessage:" + tc.fieldNum + "\"`" + `
}
`
			info, err := parseTestStruct(t, "Chat", source)
			if err != nil {
				t.Fatalf("expected valid tag, got error: %v", err)
			}

			if len(info.Fields[0].OneofVariants) != 1 {
				t.Fatal("expected 1 variant")
			}

			if info.Fields[0].OneofVariants[0].FieldNum != tc.want {
				t.Errorf("field num: got %d, want %d", info.Fields[0].OneofVariants[0].FieldNum, tc.want)
			}
		})
	}
}

func TestInterfaceRejection_AnyKeyword(t *testing.T) {
	// Test that the 'any' keyword (alias for interface{}) is rejected
	source := `
type Chat struct {
	Content any ` + "`protobuf:\"1\"`" + `
}
`
	_, err := parseTestStruct(t, "Chat", source)
	if err == nil {
		t.Error("expected error for 'any' field without oneof tag")
	}
	if !strings.Contains(err.Error(), "interface types are not supported") {
		t.Errorf("expected 'interface types are not supported' error, got: %v", err)
	}
}

func TestInterfaceRejection_InlineInterface(t *testing.T) {
	// Test that inline interface{} is rejected
	source := `
type Chat struct {
	Content interface{} ` + "`protobuf:\"1\"`" + `
}
`
	_, err := parseTestStruct(t, "Chat", source)
	if err == nil {
		t.Error("expected error for inline interface{} field without oneof tag")
	}
	if !strings.Contains(err.Error(), "interface types are not supported") {
		t.Errorf("expected 'interface types are not supported' error, got: %v", err)
	}
}

// Note: Named interface types (like `type Message interface{}`) are NOT detected
// at parse time because the AST only sees the identifier, not the underlying type.
// Users must use the `oneof` tag for polymorphic interface fields.
// Example: `Content Message `protobuf:"oneof,TextMessage:1,ImageMessage:2"``

func TestOneofTagValidation_InvalidFieldTypes(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name: "primitive string",
			source: `type Chat struct {
	Content string ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
			wantErr: "primitive type",
		},
		{
			name: "primitive int",
			source: `type Chat struct {
	Content int ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
			wantErr: "primitive type",
		},
		{
			name: "primitive bool",
			source: `type Chat struct {
	Content bool ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
			wantErr: "primitive type",
		},
		{
			name: "slice type",
			source: `type Chat struct {
	Content []string ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
			wantErr: "slice/array type",
		},
		{
			name: "map type",
			source: `type Chat struct {
	Content map[string]int ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
			wantErr: "map type",
		},
		{
			name: "pointer type",
			source: `type Message interface{}
type Chat struct {
	Content *Message ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
			wantErr: "pointer type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseTestStruct(t, "Chat", tc.source)
			if err == nil {
				t.Error("expected error for invalid oneof field type")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestOneofTagValidation_ValidFieldTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "named interface type",
			source: `type Message interface{}
type Chat struct {
	Content Message ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
		},
		{
			name: "any keyword",
			source: `type Chat struct {
	Content any ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
		},
		{
			name: "inline interface",
			source: `type Chat struct {
	Content interface{} ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
		},
		{
			name: "interface with methods",
			source: `type Chat struct {
	Content interface{ Method() } ` + "`protobuf:\"oneof,TextMessage:1\"`" + `
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := parseTestStruct(t, "Chat", tc.source)
			if err != nil {
				t.Fatalf("expected valid oneof field type, got error: %v", err)
			}
			if !info.Fields[0].IsOneof {
				t.Error("expected IsOneof to be true")
			}
		})
	}
}

func TestZeroValue(t *testing.T) {
	// zeroValue uses *new(T) for all types, which correctly returns the zero value
	tests := []string{
		"string", "bool", "int", "int32", "int64", "uint32", "float64",
		"[]byte", "*MyType", "[]MyType", "map[string]int",
		"MyStruct", "MyInt", "pkg.ExternalType",
	}

	for _, goType := range tests {
		t.Run(goType, func(t *testing.T) {
			expected := fmt.Sprintf("*new(%s)", goType)
			result := zeroValue(goType)
			if result != expected {
				t.Errorf("zeroValue(%q) = %q, want %q", goType, result, expected)
			}
		})
	}
}
