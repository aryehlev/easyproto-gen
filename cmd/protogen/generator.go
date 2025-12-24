package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/proto.tmpl
var protoTemplate string

func generateCode(buf *bytes.Buffer, pkgName string, typeNames []string, typeInfos map[string]*TypeInfo, skipHeader bool) error {
	funcMap := template.FuncMap{
		"appendFunc":        appendFunc,
		"readFunc":          readFunc,
		"unpackFunc":        unpackFunc,
		"zeroValue":         zeroValue,
		"isLengthDelimited": isLengthDelimited,
		"trimPrefix":        strings.TrimPrefix,
	}

	tmpl, err := template.New("proto").Funcs(funcMap).Parse(protoTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Package    string
		Types      []string
		TypeInfos  map[string]*TypeInfo
		SkipHeader bool
	}{
		Package:    pkgName,
		Types:      typeNames,
		TypeInfos:  typeInfos,
		SkipHeader: skipHeader,
	}

	return tmpl.Execute(buf, data)
}

// isLengthDelimited returns true for types that are length-delimited (not packed).
func isLengthDelimited(protoType string) bool {
	return protoType == "string" || protoType == "bytes"
}
