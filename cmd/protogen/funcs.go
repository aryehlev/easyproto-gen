package main

import "fmt"

// appendFunc returns the MessageMarshaler append function name for a protobuf type.
func appendFunc(protoType string, isRepeated bool) string {
	if isRepeated {
		switch protoType {
		case "int32":
			return "AppendInt32s"
		case "int64":
			return "AppendInt64s"
		case "uint32":
			return "AppendUint32s"
		case "uint64":
			return "AppendUint64s"
		case "sint32":
			return "AppendSint32s"
		case "sint64":
			return "AppendSint64s"
		case "bool":
			return "AppendBools"
		case "double":
			return "AppendDoubles"
		case "float":
			return "AppendFloats"
		case "fixed32":
			return "AppendFixed32s"
		case "fixed64":
			return "AppendFixed64s"
		case "sfixed32":
			return "AppendSfixed32s"
		case "sfixed64":
			return "AppendSfixed64s"
		}
	}

	switch protoType {
	case "string":
		return "AppendString"
	case "bytes":
		return "AppendBytes"
	case "int32", "enum":
		return "AppendInt32"
	case "int64":
		return "AppendInt64"
	case "uint32":
		return "AppendUint32"
	case "uint64":
		return "AppendUint64"
	case "sint32":
		return "AppendSint32"
	case "sint64":
		return "AppendSint64"
	case "bool":
		return "AppendBool"
	case "double":
		return "AppendDouble"
	case "float":
		return "AppendFloat"
	case "fixed32":
		return "AppendFixed32"
	case "fixed64":
		return "AppendFixed64"
	case "sfixed32":
		return "AppendSfixed32"
	case "sfixed64":
		return "AppendSfixed64"
	default:
		return "AppendBytes"
	}
}

// readFunc returns the FieldContext read function name for a protobuf type.
func readFunc(protoType string) string {
	switch protoType {
	case "string":
		return "String"
	case "bytes":
		return "Bytes"
	case "int32", "enum":
		return "Int32"
	case "int64":
		return "Int64"
	case "uint32":
		return "Uint32"
	case "uint64":
		return "Uint64"
	case "sint32":
		return "Sint32"
	case "sint64":
		return "Sint64"
	case "bool":
		return "Bool"
	case "double":
		return "Double"
	case "float":
		return "Float"
	case "fixed32":
		return "Fixed32"
	case "fixed64":
		return "Fixed64"
	case "sfixed32":
		return "Sfixed32"
	case "sfixed64":
		return "Sfixed64"
	default:
		return "Bytes"
	}
}

// unpackFunc returns the FieldContext unpack function name for packed repeated fields.
func unpackFunc(protoType string) string {
	switch protoType {
	case "int32", "enum":
		return "UnpackInt32s"
	case "int64":
		return "UnpackInt64s"
	case "uint32":
		return "UnpackUint32s"
	case "uint64":
		return "UnpackUint64s"
	case "sint32":
		return "UnpackSint32s"
	case "sint64":
		return "UnpackSint64s"
	case "bool":
		return "UnpackBools"
	case "double":
		return "UnpackDoubles"
	case "float":
		return "UnpackFloats"
	case "fixed32":
		return "UnpackFixed32s"
	case "fixed64":
		return "UnpackFixed64s"
	case "sfixed32":
		return "UnpackSfixed32s"
	case "sfixed64":
		return "UnpackSfixed64s"
	default:
		return "Bytes"
	}
}

// zeroValue returns the zero value literal for a Go type.
func zeroValue(goType string) string {
	return fmt.Sprintf("*new(%s)", goType)
}
