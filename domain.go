package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type DomainFieldGetter struct {
	Recv   string
	Name   string
	Result string
}

type DomainField struct {
	Name          string
	Type          string
	IsTypePointer bool
	IsSlice       bool
}

type DomainType struct {
	Type    string
	File    string
	Fields  []*DomainField
	Getters []DomainFieldGetter
}

type DTOInitFunc struct {
	Func   string
	Params []string
}

type DTOField struct {
	Name string
	Type string
}

type DTOType struct {
	Type               string
	IsAggregateRoot    bool
	MapToDomain        bool
	MapFromDomain      bool
	IgnoreDomainFields []string
	Fields             []DTOField
}

type DomainConfig struct {
	Dir         string
	Package     string
	Filename    string
	DomainTypes []DomainType
	DTOSuffix   string
	MapFromFunc string
	MapToFunc   string
}

func IsPointer(field string) bool {
	return strings.HasPrefix(field, "*")
}

func IsSlice(field string) bool {
	return strings.HasPrefix(field, "[]")
}

func IsSliceOfPointers(field string) bool {
	return strings.Contains(field, "[]*")
}

func UnmarshalDomainConfigYaml(filename string) error {

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	m := make(map[string]interface{})

	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	// fmt.Printf("--- m:\n%v\n\n", m)

	// fmt.Printf("--- domain_types:\n%v\n\n", m["domain_types"])

	dtypes := make(map[string][]string)

	for k, v := range (m["domain_types"]).(map[interface{}]interface{}) {

		mk := k.(string)
		mv := v.(map[interface{}]interface{})

		// fmt.Println(mk, mv)

		for k1, v1 := range mv {

			mk1 := k1.(string)
			mv1 := v1.(string)

			if mk1 == "file" {
				dtypes[mv1] = append(dtypes[mv1], mk)
			}

		}

	}

	domainTypes := make(map[string]*DomainType)

	for file, types := range dtypes {
		findDomainTypeFields(file, types, domainTypes)
	}

	for _, domainType := range domainTypes {
		fmt.Println(domainType.Type, domainType.File)

		for _, field := range domainType.Fields {
			fmt.Println("\t", field)
		}

		for _, getter := range domainType.Getters {
			fmt.Println("\t", getter)
		}

	}

	return nil
}

func findDomainTypeFields(file string, domainTypes []string, dtypes map[string]*DomainType) ([]DomainType, error) {

	src, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Base(file), src, 0)
	if err != nil {
		return nil, err
	}

	for _, dtype := range domainTypes {
		dtypes[dtype] = &DomainType{
			File: file,
			Type: dtype,
		}
	}

	ast.Inspect(f, func(n ast.Node) bool {

		parseTree(n, dtypes)

		return true
	})

	return nil, nil
}

func fieldTypeParser(ftype ast.Expr) string {

	var fieldType string

	switch fi := ftype.(type) {
	case *ast.Ident:
		fieldType = fi.Name
	case *ast.SelectorExpr:
		fieldType = fi.Sel.Name
	case *ast.StarExpr:
		fieldType = fmt.Sprintf("*%s", fi.X.(*ast.Ident).Name)
	case *ast.ArrayType:
		fieldType = fmt.Sprintf("[]%s", fieldTypeParser(fi.Elt))
	case *ast.MapType:
		break
	default:
		fmt.Printf("Handle this type: %T\n", fi)
	}

	return fieldType
}

func structTypeParser(strctype *ast.StructType) []*DomainField {

	fields := make([]*DomainField, strctype.Fields.NumFields())

	for index, field := range strctype.Fields.List {

		fieldType := fieldTypeParser(field.Type)
		var fieldName string

		for _, ident := range field.Names {
			if ident.Name != "" {
				fieldName = ident.Name
				break
			}
		}

		isSlicePointers := IsSliceOfPointers(fieldType)

		fields[index] = &DomainField{
			Name:          fieldName,
			Type:          fieldType,
			IsTypePointer: isSlicePointers || IsPointer(fieldType),
			IsSlice:       isSlicePointers || IsSlice(fieldType),
		}

	}

	return fields
}

func parseTree(n ast.Node, dtypes map[string]*DomainType) {

	switch x := n.(type) {
	case *ast.TypeSpec:
		switch types := x.Type.(type) {
		case *ast.StructType:
			if dtypes[x.Name.Name] == nil {
				break
			}
			// fmt.Println(x.Name.Name)
			dtypes[x.Name.Name].Fields = structTypeParser(types)
		default:
			fmt.Println(x.Name.Name)
		}
	case *ast.FuncDecl:

		if x.Recv != nil {

			var recv string

			for _, field := range x.Recv.List {
				if field == nil {
					continue
				}
				recv = fieldTypeParser(field.Type)
				break
			}

			dtype := dtypes[recv]
			if dtype == nil {
				break
			}

			funcName := fieldTypeParser(x.Name)
			var resstr string

			if x.Type.Results != nil {
				for _, res := range x.Type.Results.List {
					resstr = fieldTypeParser(res.Type)
				}
			}

			getter := DomainFieldGetter{
				Recv:   recv,
				Name:   funcName,
				Result: resstr,
			}

			dtypes[recv].Getters = append(dtypes[recv].Getters, getter)
		}

	}

}
