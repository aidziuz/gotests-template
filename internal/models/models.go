package models

import (
	"strings"
	"unicode"
)

type Expression struct {
	Value      string
	IsStar     bool
	IsVariadic bool
	IsWriter   bool
	Underlying string
}

func (e *Expression) String() string {
	value := e.Value
	if e.IsStar {
		value = "*" + value
	}
	if e.IsVariadic {
		return "[]" + value
	}
	return value
}

type Field struct {
	Name  string
	Type  *Expression
	Index int
}

func (f *Field) IsWriter() bool {
	return f.Type.IsWriter
}

func (f *Field) IsStruct() bool {
	return strings.HasPrefix(f.Type.Underlying, "struct")
}

func (f *Field) IsBasicType() bool {
	return isBasicType(f.Type.String()) || isBasicType(f.Type.Underlying)
}

func (f *Field) GetBasicValue() string {
	if isBasicType(f.Type.String()) {
		return getBasicValue(f.Type.String())
	}

	if isBasicType(f.Type.Underlying) {
		return getBasicValue(f.Type.Underlying)
	}

	return "nil"
}

func (f *Field) HasNoPackage() bool {
	return !isBasicType(f.Type.String()) && !strings.Contains(f.Type.Value, ".")
}

func isBasicType(t string) bool {
	switch t {
	case "bool", "string", "int", "int8", "int16", "int32", "int64", "uint",
		"uint8", "uint16", "uint32", "uint64", "uintptr", "byte", "rune",
		"float32", "float64", "complex64", "complex128":
		return true
	default:
		if strings.HasPrefix(t, "map[string]") {
			return true
		}
		return false
	}
}

func getBasicValue(t string) string {
	switch t {
	case "bool":
		return "false"
	case "string":
		return "\"stringValue\""
	default:
		return "0"
	}
}

func (f *Field) IsNamed() bool {
	return f.Name != "" && f.Name != "_"
}

func (f *Field) ShortName() string {
	return strings.ToLower(string([]rune(f.Type.Value)[0]))
}

type Receiver struct {
	*Field
	Fields []*Field
}

type Function struct {
	Name          string
	IsExported    bool
	IsConstructor bool
	Receiver      *Receiver
	Parameters    []*Field
	Results       []*Field
	ReturnsError  bool
}

func (f *Function) TestParameters() []*Field {
	var ps []*Field
	for _, p := range f.Parameters {
		if p.IsWriter() {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}

func (f *Function) TestResults() []*Field {
	var ps []*Field
	ps = append(ps, f.Results...)
	for _, p := range f.Parameters {
		if !p.IsWriter() {
			continue
		}
		ps = append(ps, &Field{
			Name: p.Name,
			Type: &Expression{
				Value:      "string",
				IsWriter:   true,
				Underlying: "string",
			},
			Index: len(ps),
		})
	}
	return ps
}

func (f *Function) ReturnsMultiple() bool {
	return len(f.Results) > 1
}

func (f *Function) OnlyReturnsOneValue() bool {
	return len(f.Results) == 1 && !f.ReturnsError
}

func (f *Function) OnlyReturnsError() bool {
	return len(f.Results) == 0 && f.ReturnsError
}

func (f *Function) FullName() string {
	var r string
	if f.Receiver != nil {
		r = f.Receiver.Type.Value
	}
	return strings.Title(r) + strings.Title(f.Name)
}

func (f *Function) TestName() string {
	if strings.HasPrefix(f.Name, "Test") {
		return f.Name
	}
	if f.Receiver != nil {
		receiverType := f.Receiver.Type.Value
		if unicode.IsLower([]rune(receiverType)[0]) {
			receiverType = "_" + receiverType
		}
		return "Test" + receiverType + "_" + f.Name
	}
	if unicode.IsLower([]rune(f.Name)[0]) {
		return "Test_" + f.Name
	}
	return "Test" + f.Name
}

func (f *Function) IsNaked() bool {
	return f.Receiver == nil && len(f.Parameters) == 0 && len(f.Results) == 0
}

type Import struct {
	Name, Path string
}

type Header struct {
	Comments    []string
	Package     string
	TestPackage string
	Imports     []*Import
	Code        []byte
}

type Path string

func (p Path) TestPath() string {
	if !p.IsTestPath() {
		return strings.TrimSuffix(string(p), ".go") + "_test.go"
	}
	return string(p)
}

func (p Path) IsTestPath() bool {
	return strings.HasSuffix(string(p), "_test.go")
}
