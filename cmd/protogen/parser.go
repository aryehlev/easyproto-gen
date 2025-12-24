package main

import (
	"fmt"
	"go/ast"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func parseStruct(typeName string, structType *ast.StructType) (*TypeInfo, error) {
	info := &TypeInfo{
		Name: typeName,
	}

	// Track field numbers to detect duplicates
	seenFieldNums := make(map[int]string)

	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}

		// Parse the struct tag
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		protoTag := tag.Get("protobuf")
		if protoTag == "" {
			continue
		}

		parts := strings.Split(protoTag, ",")

		// Check for oneof tag format: `protobuf:"oneof,TypeA:1,TypeB:2"`
		isOneof := strings.TrimSpace(parts[0]) == "oneof"
		var oneofVariants []OneofVariant
		var fieldNum int
		var err error

		if isOneof {
			// Validate that the field type is valid for oneof (must be interface-like, not primitive/slice/map)
			if err := validateOneofFieldType(field.Type); err != nil {
				fieldName := ""
				if len(field.Names) > 0 {
					fieldName = field.Names[0].Name
				}
				return nil, fmt.Errorf("invalid oneof field %q in type %s: %w", fieldName, typeName, err)
			}

			// Parse oneof variants
			if len(parts) < 2 {
				return nil, fmt.Errorf("oneof tag requires at least one variant: %s", protoTag)
			}
			for _, part := range parts[1:] {
				part = strings.TrimSpace(part)
				colonIdx := strings.LastIndex(part, ":")
				if colonIdx == -1 {
					return nil, fmt.Errorf("invalid oneof variant %q in tag %q: expected Type:FieldNum format", part, protoTag)
				}
				variantType := strings.TrimSpace(part[:colonIdx])
				variantFieldNum, err := strconv.Atoi(strings.TrimSpace(part[colonIdx+1:]))
				if err != nil {
					return nil, fmt.Errorf("invalid field number for oneof variant %q in tag %q", part, protoTag)
				}
				// Validate variant field number
				if variantFieldNum < 1 || variantFieldNum > 536870911 {
					return nil, fmt.Errorf("invalid field number %d for oneof variant %q: must be 1-536870911", variantFieldNum, variantType)
				}
				if variantFieldNum >= 19000 && variantFieldNum <= 19999 {
					return nil, fmt.Errorf("invalid field number %d for oneof variant %q: range 19000-19999 is reserved", variantFieldNum, variantType)
				}
				// Check for duplicate field numbers in oneof
				for _, existing := range oneofVariants {
					if existing.FieldNum == variantFieldNum {
						return nil, fmt.Errorf("duplicate field number %d in oneof: used by both %q and %q", variantFieldNum, existing.TypeName, variantType)
					}
				}
				oneofVariants = append(oneofVariants, OneofVariant{
					TypeName: variantType,
					FieldNum: variantFieldNum,
				})
			}
			// Use -1 as sentinel for oneof (no single field number)
			fieldNum = -1
		} else {
			fieldNum, err = strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid field number in tag %q: must be a number", protoTag)
			}

			// Validate field number range (protobuf spec: 1 to 2^29-1, with 19000-19999 reserved)
			if fieldNum < 1 {
				return nil, fmt.Errorf("invalid field number %d in tag %q: must be >= 1", fieldNum, protoTag)
			}
			if fieldNum > 536870911 { // 2^29 - 1
				return nil, fmt.Errorf("invalid field number %d in tag %q: must be <= 536870911", fieldNum, protoTag)
			}
			if fieldNum >= 19000 && fieldNum <= 19999 {
				return nil, fmt.Errorf("invalid field number %d in tag %q: range 19000-19999 is reserved", fieldNum, protoTag)
			}
		}

		// Proto type is optional - can be inferred from Go type
		var protoType string
		if isOneof {
			protoType = "oneof"
		} else if len(parts) >= 2 {
			protoType = strings.TrimSpace(parts[1])
			// Validate explicit protobuf type
			if !isValidProtoType(protoType) {
				return nil, fmt.Errorf("invalid protobuf type %q in tag %q", protoType, protoTag)
			}
		} else {
			// Infer from Go type
			protoType = inferProtoType(field.Type)
		}

		// Reject interface types (like 'any' or custom interfaces)
		// Note: oneof fields never reach here with protoType="interface" since they get protoType="oneof" above
		if protoType == "interface" {
			fieldName := ""
			if len(field.Names) > 0 {
				fieldName = field.Names[0].Name
			}
			return nil, fmt.Errorf("interface types are not supported for protobuf (use oneof tag for polymorphism): field %q in type %s has type %s",
				fieldName, typeName, exprToString(field.Type))
		}

		// Check for options
		isRepeated := false
		isOptional := false
		isEnum := protoType == "enum"
		isMap := protoType == "map"
		isCustom := false

		// For maps, we need key and value types from the tag or infer them
		var mapKeyProto, mapValueProto string
		var mapValueCustom bool
		if isMap {
			if len(parts) >= 4 {
				// Explicit: `protobuf:"1,map,string,int32"`
				mapKeyProto = strings.TrimSpace(parts[2])
				mapValueProto = strings.TrimSpace(parts[3])
			} else if mapType, ok := field.Type.(*ast.MapType); ok {
				// Infer from Go type: `protobuf:"1"` on map[string]int32
				mapKeyProto = inferProtoType(mapType.Key)
				mapValueProto = inferProtoType(mapType.Value)
			}
			// Validate map key type (only certain scalar types allowed)
			if !isValidMapKeyType(mapKeyProto) {
				return nil, fmt.Errorf("invalid map key type %q in tag %q: must be string, bool, or integer type", mapKeyProto, protoTag)
			}
		}

		// Parse options from remaining parts (if any) - skip for oneof (already parsed)
		if !isOneof {
			optionStart := 2
			if isMap && len(parts) >= 4 {
				optionStart = 4 // Skip map key/value types
			}
			if len(parts) > optionStart {
				for _, part := range parts[optionStart:] {
					switch strings.TrimSpace(part) {
					case "repeated":
						isRepeated = true
					case "optional":
						isOptional = true
					case "enum":
						isEnum = true
					case "custom":
						isCustom = true
						// For maps, custom applies to the value type
						if isMap {
							mapValueCustom = true
						}
					}
				}
			}
		}

		// Handle embedded fields (anonymous fields) - they have no Names
		fieldNames := make([]string, 0, len(field.Names))
		for _, name := range field.Names {
			fieldNames = append(fieldNames, name.Name)
		}
		if len(fieldNames) == 0 {
			// Embedded field - use the type name as the field name
			embeddedName := getTypeName(field.Type)
			if embeddedName == "" {
				return nil, fmt.Errorf("cannot determine name for embedded field with tag %q in type %s", protoTag, typeName)
			}
			fieldNames = append(fieldNames, embeddedName)
		}

		for _, fieldName := range fieldNames {
			// Check for duplicate field numbers
			if isOneof {
				// For oneof, check all variant field numbers
				for _, variant := range oneofVariants {
					if existingField, ok := seenFieldNums[variant.FieldNum]; ok {
						return nil, fmt.Errorf("duplicate field number %d: used by both %q and oneof variant %q in type %s",
							variant.FieldNum, existingField, variant.TypeName, typeName)
					}
					seenFieldNums[variant.FieldNum] = fieldName + ":" + variant.TypeName
				}
			} else {
				if existingField, ok := seenFieldNums[fieldNum]; ok {
					return nil, fmt.Errorf("duplicate field number %d: used by both %q and %q in type %s",
						fieldNum, existingField, fieldName, typeName)
				}
				seenFieldNums[fieldNum] = fieldName
			}

			fi := &FieldInfo{
				Name:          fieldName,
				FieldNum:      fieldNum,
				ProtoType:     protoType,
				IsRepeated:    isRepeated,
				IsOptional:    isOptional,
				IsMessage:     protoType == "message",
				IsEnum:        isEnum,
				IsMap:         isMap,
				IsCustom:      isCustom,
				IsOneof:       isOneof,
				OneofVariants: oneofVariants,
			}

			// Analyze Go type
			fi.GoType = exprToString(field.Type)
			analyzeType(fi, field.Type)

			// Handle map-specific parsing
			if fi.IsMap {
				fi.MapKeyProto = mapKeyProto
				fi.MapValueProto = mapValueProto
				fi.MapValueIsMsg = mapValueProto == "message"
				fi.MapValueCustom = mapValueCustom
				// Extract key/value Go types from the AST
				if mapType, ok := field.Type.(*ast.MapType); ok {
					fi.MapKeyType = exprToString(mapType.Key)
					fi.MapValueType = exprToString(mapType.Value)
					// Check if value is a pointer
					if _, isPtr := mapType.Value.(*ast.StarExpr); isPtr {
						fi.MapValueIsPtr = true
					}
				}
			}

			// Handle enum type conversion
			if fi.IsEnum {
				fi.NeedsTypeConv = true
				fi.ConvType = "int32"
			}

			info.Fields = append(info.Fields, fi)
		}
	}

	// Sort fields by field number
	sort.Slice(info.Fields, func(i, j int) bool {
		return info.Fields[i].FieldNum < info.Fields[j].FieldNum
	})

	return info, nil
}

// getTypeName extracts the type name from an AST expression (for embedded fields)
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	case *ast.StarExpr:
		return getTypeName(t.X)
	default:
		return ""
	}
}

// inferProtoType infers the protobuf type from a Go AST type expression.
func inferProtoType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "bool":
			return "bool"
		case "int32":
			return "int32"
		case "int64":
			return "int64"
		case "int":
			return "int64"
		case "uint32":
			return "uint32"
		case "uint64":
			return "uint64"
		case "uint":
			return "uint64"
		case "float32":
			return "float"
		case "float64":
			return "double"
		case "byte":
			return "int32"
		case "any":
			return "interface"
		default:
			return "message"
		}
	case *ast.InterfaceType:
		return "interface"
	case *ast.SelectorExpr:
		return "message"
	case *ast.StarExpr:
		return inferProtoType(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			if ident, ok := t.Elt.(*ast.Ident); ok && ident.Name == "byte" {
				return "bytes"
			}
			return inferProtoType(t.Elt)
		}
		return inferProtoType(t.Elt)
	case *ast.MapType:
		return "map"
	default:
		return "bytes"
	}
}

func analyzeType(fi *FieldInfo, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.Ident:
		fi.BaseType = t.Name
		fi.ElemType = t.Name
		fi.RawElemType = t.Name
	case *ast.SelectorExpr:
		fullType := exprToString(t)
		fi.BaseType = fullType
		fi.ElemType = fullType
		fi.RawElemType = fullType
	case *ast.StarExpr:
		fi.IsPointer = true
		fi.IsOptional = true
		inner := exprToString(t.X)
		fi.BaseType = inner
		fi.ElemType = inner
		fi.RawElemType = inner
	case *ast.ArrayType:
		if t.Len == nil {
			// Special case: []byte is NOT a repeated field
			if ident, ok := t.Elt.(*ast.Ident); ok && ident.Name == "byte" {
				fi.BaseType = "[]byte"
				fi.ElemType = "byte"
				fi.RawElemType = "byte"
				return
			}

			fi.IsRepeated = true
			if star, ok := t.Elt.(*ast.StarExpr); ok {
				fi.IsSliceOfPtr = true
				fi.ElemType = exprToString(star.X)
				fi.RawElemType = "*" + fi.ElemType
				fi.BaseType = fi.ElemType
			} else {
				fi.ElemType = exprToString(t.Elt)
				fi.RawElemType = fi.ElemType
				fi.BaseType = fi.ElemType
			}
		}
	}
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", exprToString(t.Len), exprToString(t.Elt))
	case *ast.BasicLit:
		return t.Value
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", exprToString(t.Key), exprToString(t.Value))
	default:
		return fmt.Sprintf("%T", expr)
	}
}
