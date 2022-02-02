package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type DomainField struct {
	Name   string
	Type   string
	Getter string
}

type DomainType struct {
	Type   string
	File   string
	Fields []DomainField
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

	var domainTypes []DomainType

	for k, v := range (m["domain_types"]).(map[interface{}]interface{}) {

		mk := k.(string)
		mv := v.(map[interface{}]interface{})

		// fmt.Println(mk, mv)

		d := DomainType{
			Type: mk,
		}

		for k1, v1 := range mv {

			mk1 := k1.(string)
			mv1 := v1.(string)

			if mk1 == "file" {
				d.File = mv1
			}

		}

		domainTypes = append(domainTypes, d)
	}

	// fmt.Println(domainTypes)

	for _, dtype := range domainTypes {
		findDomainTypeFields(dtype.Type, dtype.File)
	}

	return nil
}

func findDomainTypeFields(domainType string, file string) ([]DomainField, error) {

	src, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Base(file), src, 0)
	if err != nil {
		return nil, err
	}

	ast.Inspect(f, func(n ast.Node) bool {

		parseTree(n)

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

func structTypeParser(strctype *ast.StructType) {

	for _, field := range strctype.Fields.List {

		fieldType := fieldTypeParser(field.Type)

		for _, indent := range field.Names {
			fmt.Println("\t", indent.Name, fieldType)
		}

	}

}

func parseTree(n ast.Node) {

	switch x := n.(type) {
	case *ast.TypeSpec:
		switch types := x.Type.(type) {
		case *ast.StructType:
			fmt.Println(x.Name.Name)
			structTypeParser(types)
		default:
			fmt.Println(x.Name.Name)
		}
	case *ast.FuncDecl:

		funcName := fieldTypeParser(x.Name)

		if x.Type.Results != nil {
			for _, res := range x.Type.Results.List {
				resstr := fieldTypeParser(res.Type)
				fmt.Println(resstr)
			}
		}

		if x.Recv != nil {
			for _, field := range x.Recv.List {
				if field == nil {
					continue
				}
				fieldType := fieldTypeParser(field.Type)
				fmt.Printf("(%s) %s\n", fieldType, funcName)
			}
		}

	}

}
