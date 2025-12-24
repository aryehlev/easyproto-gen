// protogen generates protobuf marshal/unmarshal functions for structs with protobuf tags.
//
// Usage:
//
//	//go:generate go run github.com/VictoriaMetrics/easyproto/cmd/protogen -type=Timeseries,Sample
//
// Struct tags format:
//
//	`protobuf:"fieldNum[,type][,options...]"`
//
// The type is OPTIONAL - it will be inferred from the Go type when omitted:
//
//	string    -> string       int32   -> int32      float32 -> float
//	[]byte    -> bytes        int64   -> int64      float64 -> double
//	bool      -> bool         uint32  -> uint32     CustomType -> message
//	int       -> int64        uint64  -> uint64     map[K]V -> map
//
// Options:
//   - repeated: field is a repeated (slice) field
//   - optional: field is optional (pointer type, nil means unset)
//   - enum: field is an enum type (uses int32 wire type)
//
// When you need non-default wire types, specify explicitly:
//   - sint32, sint64: for signed integers with many negative values
//   - fixed32, fixed64, sfixed32, sfixed64: for fixed-width encoding
//
// Example with inferred types (simple):
//
//	type Timeseries struct {
//	    Name    string   `protobuf:"1"`          // inferred: string
//	    Samples []Sample `protobuf:"2"`          // inferred: message (repeated)
//	}
//
//	type Sample struct {
//	    Value     float64 `protobuf:"1"`         // inferred: double
//	    Timestamp int64   `protobuf:"2"`         // inferred: int64
//	}
//
//	type WithMaps struct {
//	    Data  map[string]int32   `protobuf:"1"` // inferred: map<string,int32>
//	    Items map[string]*Sample `protobuf:"2"` // inferred: map<string,message>
//	}
//
// Example with explicit types (when needed):
//
//	type Explicit struct {
//	    SignedVal int32  `protobuf:"1,sint32"`   // use sint32 encoding
//	    FixedVal  uint64 `protobuf:"2,fixed64"`  // use fixed64 encoding
//	    Status    MyEnum `protobuf:"3,enum"`     // enum type
//	}
//
// Oneof fields (polymorphic interfaces):
//
//	type Message interface { MessageType() string }
//	type TextMessage struct { Text string `protobuf:"1"` }
//	type ImageMessage struct { URL string `protobuf:"1"` }
//
//	type Chat struct {
//	    Content Message `protobuf:"oneof,TextMessage:1,ImageMessage:2"`
//	}
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	typeNames = flag.String("type", "", "comma-separated list of type names")
	output    = flag.String("output", "", "output file name; default srcdir/<type>_proto.go")
	noHeader  = flag.Bool("noheader", false, "skip generating the _mp pool and interface definitions (use when adding to existing generated file)")
)

func main() {
	flag.Parse()

	if *typeNames == "" {
		log.Fatal("-type flag is required")
	}

	types := strings.Split(*typeNames, ",")
	for i := range types {
		types[i] = strings.TrimSpace(types[i])
	}

	// Get the directory to parse
	dir := "."
	if len(flag.Args()) > 0 {
		dir = flag.Args()[0]
	}

	// Parse .go files in directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read directory %s: %v", dir, err)
	}

	fset := token.NewFileSet()
	var files []*ast.File
	var pkgName string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		filePath := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("failed to parse file %s: %v", filePath, err)
		}
		if pkgName == "" {
			pkgName = file.Name.Name
		} else if file.Name.Name != pkgName {
			continue // skip files from different packages
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		log.Fatal("no Go files found")
	}

	// Find the requested types
	typeInfos := make(map[string]*TypeInfo)
	for _, file := range files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				for _, typeName := range types {
					if typeSpec.Name.Name == typeName {
						structType, ok := typeSpec.Type.(*ast.StructType)
						if !ok {
							log.Fatalf("type %s is not a struct", typeName)
						}
						info, err := parseStruct(typeName, structType)
						if err != nil {
							log.Fatalf("failed to parse struct %s: %v", typeName, err)
						}
						typeInfos[typeName] = info
					}
				}
			}
		}
	}

	// Check all types were found
	for _, typeName := range types {
		if _, ok := typeInfos[typeName]; !ok {
			log.Fatalf("type %s not found", typeName)
		}
	}

	// Generate code
	var buf bytes.Buffer
	if err := generateCode(&buf, pkgName, types, typeInfos, *noHeader); err != nil {
		log.Fatalf("failed to generate code: %v", err)
	}

	// Format the code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		tmpFile, tmpErr := os.CreateTemp("", "protogen_debug_*.go")
		if tmpErr == nil {
			tmpFile.Write(buf.Bytes())
			tmpFile.Close()
			log.Fatalf("failed to format generated code (debug output: %s): %v", tmpFile.Name(), err)
		}
		log.Fatalf("failed to format generated code: %v", err)
	}

	// Determine output file
	outputFile := *output
	if outputFile == "" {
		if len(types) == 1 {
			outputFile = filepath.Join(dir, strings.ToLower(types[0])+"_proto.go")
		} else {
			outputFile = filepath.Join(dir, pkgName+"_proto.go")
		}
	}

	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}

	fmt.Printf("Generated %s\n", outputFile)
}
