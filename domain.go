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
	Recv       string
	Name       string
	ResultType FieldType
}

type DomainFunc struct {
	Recv       string
	Name       string
	ResultType FieldType
}

type FieldType struct {
	Package       string
	Type          string
	IsTypePointer bool
	IsSlice       bool
}

type DomainField struct {
	Name   string
	IsMap  bool
	MapKey FieldType
	Type   FieldType
	Getter DomainFieldGetter
}

type DomainType struct {
	Type   string
	File   string
	Fields map[string]*DomainField
	Funcs  []DomainFunc
}

type DTOField struct {
	Name      string
	MappingTo string
	IsMap     bool
	MapKey    FieldType
	Type      FieldType
}

type DTOMapBinding struct {
	Name  string
	Type  string
	Field string
}

type DTOType struct {
	Type            string
	IsAggregateRoot bool
	IsDomainMapping bool

	IgnoreDomainFields []string
	Fields             map[string]*DTOField

	MapBindings map[string]DTOMapBinding
	Func        string
	Params      []string
}

type DomainConfig struct {
	Dir         string
	Package     string
	Filename    string
	DomainTypes map[string]*DomainType
	DTOTypes    map[string]*DTOType
	DTOSuffix   string
	MapFromFunc string
	MapToFunc   string
	Imports     map[string]string
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
	// fmt.Printf("--- domain_dto_types:\n%v\n\n", m["domain_dto_types"])

	dc := &DomainConfig{}

	dc.Dir = m["dir"].(string)
	dc.Package = m["package"].(string)
	dc.Filename = m["filename"].(string)

	dc.Imports = make(map[string]string)

	for k, v := range m["imports"].(map[interface{}]interface{}) {
		imppkg := k.(string)
		url := v.(string)
		dc.Imports[imppkg] = url
	}

	dc.DomainTypes = make(map[string]*DomainType)
	dtypes := parseDomainTypes(m["domain_types"].(map[interface{}]interface{}))
	for file, types := range dtypes {
		findDomainTypeFields(file, types, dc.DomainTypes)
	}

	dc.DTOSuffix = m["dto_suffix"].(string)
	dc.MapFromFunc = m["map_from_func"].(string)
	dc.MapToFunc = m["map_to_func"].(string)
	dc.DTOTypes = parseDTOTypes(m["domain_dto_types"].(map[interface{}]interface{}))

	clean(dc)

	// for _, domainType := range domainTypes {
	// 	fmt.Println(domainType.Type, domainType.File)

	// 	for _, field := range domainType.Fields {
	// 		fmt.Println("\t", field.Name, field.Type, field.Type.IsSlice, field.Type.IsTypePointer)
	// 		fmt.Println("\t\t", field.Getter)

	// 	}

	// }

	GenerateDomainMappers(dc)

	return nil
}

func parseDTOTypes(m map[interface{}]interface{}) map[string]*DTOType {

	dtotypes := make(map[string]*DTOType)

	for k, v := range m {

		mk := k.(string)
		mv := v.(map[interface{}]interface{})

		dto := &DTOType{}
		dto.Fields = make(map[string]*DTOField)
		dto.MapBindings = make(map[string]DTOMapBinding)
		dto.Type = mk
		dto.IsAggregateRoot = mv["is_aggregate_root"].(bool)
		if dto.IsAggregateRoot {
			dto.IsDomainMapping = mv["domain_mapping_enabled"].(bool)
		}

		if mv["ignore_domain_fields"] != nil {
			ignoreFields := mv["ignore_domain_fields"].([]interface{})
			dto.IgnoreDomainFields = make([]string, len(ignoreFields))
			for index, fi := range ignoreFields {
				dto.IgnoreDomainFields[index] = fi.(string)
			}
		}

		if mv["map_bindings"] != nil {
			mapBindings := mv["map_bindings"].(map[interface{}]interface{})

			for mKey, mVal := range mapBindings {
				for _, m2Val := range mVal.(map[interface{}]interface{}) {

					expStr := strings.Split(m2Val.(string), ".")

					m1Key := mKey.(string)

					dto.MapBindings[m1Key] = DTOMapBinding{
						Name:  m1Key,
						Type:  expStr[0],
						Field: expStr[1],
					}

					// fmt.Println(dto.MapBindings[m1Key])
				}
			}
		}

		for k2, v2 := range mv {

			switch mv2 := v2.(type) {
			case map[interface{}]interface{}:

				if k2.(string) == "map_bindings" {
					continue
				}

				if mv2["is_init_func"].(bool) {
					dto.Func = k2.(string)

					params := mv2["param_mapping"].([]interface{})
					dto.Params = make([]string, len(params))
					for i, param := range params {
						dto.Params[i] = param.(string)
					}
				}

			}
		}

		dtotypes[mk] = dto
	}

	return dtotypes
}

func parseDomainTypes(m map[interface{}]interface{}) map[string][]string {

	dtypes := make(map[string][]string)

	for k, v := range m {

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

	return dtypes
}

func findIgnoredKeyboard(unexFunc string, dto *DTOType) bool {
	for _, ignored := range dto.IgnoreDomainFields {
		if ignored == unexFunc {
			return true
		}
	}

	return false
}

func clean(dc *DomainConfig) {

	// 1. check for getters

	for dtypestr, dtype := range dc.DomainTypes {
		dto := dc.DTOTypes[dtypestr]

		for dname, dfield := range dtype.Fields {

			if url, ok := dc.Imports[dfield.MapKey.Package]; ok {
				dfield.MapKey.Package = url
			}

			if url, ok := dc.Imports[dfield.Type.Package]; ok {
				dfield.Type.Package = url
			}

			if findIgnoredKeyboard(dname, dto) {

				dc.DomainTypes[dtypestr].Fields[dname] = &DomainField{
					Name:   dfield.Name,
					IsMap:  dfield.IsMap,
					MapKey: dfield.MapKey,
					Type:   dfield.Type,
					Getter: dfield.Getter,
				}

				continue
			}

			dfield.Getter = DomainFieldGetter{
				Recv: dtypestr,
				Name: dfield.Name,
			}

			name := strings.Title(dfield.Name)

			if dfield.IsMap {
				fmt.Println(dfield.Name, dfield.IsMap, dfield.MapKey, dfield.Type)
			}

			dto.Fields[name] = &DTOField{
				Name:      name,
				MappingTo: dfield.Name,
				IsMap:     dfield.IsMap,
				MapKey:    dfield.MapKey,
				Type:      dfield.Type,
			}

		}

		dtype.Funcs = nil
	}

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
			File:   file,
			Type:   dtype,
			Fields: make(map[string]*DomainField),
		}
	}

	ast.Inspect(f, func(n ast.Node) bool {

		parseTree(n, dtypes)

		return true
	})

	return nil, nil
}
