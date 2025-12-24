package main

import (
	"fmt"
	"go/ast"
)

// validProtoTypes is the set of valid protobuf types
var validProtoTypes = map[string]bool{
	"string":   true,
	"bytes":    true,
	"int32":    true,
	"int64":    true,
	"uint32":   true,
	"uint64":   true,
	"sint32":   true,
	"sint64":   true,
	"bool":     true,
	"double":   true,
	"float":    true,
	"fixed32":  true,
	"fixed64":  true,
	"sfixed32": true,
	"sfixed64": true,
	"message":  true,
	"enum":     true,
	"map":      true,
	"oneof":    true,
}

// validMapKeyTypes is the set of valid protobuf map key types
var validMapKeyTypes = map[string]bool{
	"string":   true,
	"int32":    true,
	"int64":    true,
	"uint32":   true,
	"uint64":   true,
	"sint32":   true,
	"sint64":   true,
	"fixed32":  true,
	"fixed64":  true,
	"sfixed32": true,
	"sfixed64": true,
	"bool":     true,
	// Note: float/double and bytes are NOT valid map key types in protobuf
}

// isValidProtoType checks if a protobuf type is valid
func isValidProtoType(protoType string) bool {
	return validProtoTypes[protoType]
}

// isValidMapKeyType checks if a protobuf type is valid as a map key
func isValidMapKeyType(protoType string) bool {
	return validMapKeyTypes[protoType]
}

// validateOneofFieldType checks if a field type is valid for oneof usage.
// Oneof fields must be interface types (named or inline), not primitives, slices, or maps.
func validateOneofFieldType(expr ast.Expr) error {
	switch t := expr.(type) {
	case *ast.Ident:
		// Check for primitive types that can't be oneof
		switch t.Name {
		case "string", "bool", "byte", "rune",
			"int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64", "complex64", "complex128",
			"uintptr":
			return fmt.Errorf("oneof tag cannot be used on primitive type %q", t.Name)
		case "any":
			// 'any' is valid for oneof
			return nil
		default:
			// Named type (could be interface or struct) - we allow it
			// At runtime, if it's not an interface, the type switch will fail
			return nil
		}
	case *ast.InterfaceType:
		// Inline interface{} or interface{...} - valid for oneof
		return nil
	case *ast.ArrayType:
		return fmt.Errorf("oneof tag cannot be used on slice/array type %s", exprToString(expr))
	case *ast.MapType:
		return fmt.Errorf("oneof tag cannot be used on map type %s", exprToString(expr))
	case *ast.StarExpr:
		return fmt.Errorf("oneof tag cannot be used on pointer type %s (the variants are stored as pointers)", exprToString(expr))
	case *ast.ChanType:
		return fmt.Errorf("oneof tag cannot be used on channel type %s", exprToString(expr))
	case *ast.FuncType:
		return fmt.Errorf("oneof tag cannot be used on function type")
	case *ast.SelectorExpr:
		// Qualified type like pkg.Type - could be interface, allow it
		return nil
	default:
		return fmt.Errorf("oneof tag cannot be used on type %s", exprToString(expr))
	}
}
