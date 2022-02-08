package main

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"
)

func IsPointer(field string) bool {
	return strings.HasPrefix(field, "*")
}

func IsSlice(field string) bool {
	return strings.HasPrefix(field, "[]")
}

func IsSliceOfPointers(field string) bool {
	return strings.Contains(field, "[]*")
}

func RemovePointer(field string) string {
	return strings.Replace(field, "*", "", 1)
}

func RemoveArray(field string) string {
	return strings.Replace(field, "[]", "", 1)
}

func LowercaseFirstLetter(str string) string {
	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func UppercaseFirstLetter(str string) string {
	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func fieldTypeParser(ftype ast.Expr) string {

	var fieldType string

	switch fi := ftype.(type) {
	case *ast.Ident:
		fieldType = fi.Name
	case *ast.SelectorExpr:
		fldtype := fi.Sel.Name
		pkg := fieldTypeParser(fi.X)
		if pkg != "" {
			fldtype = fmt.Sprintf("%s.%s", pkg, fi.Sel.Name)
		}
		fieldType = fldtype
	case *ast.StarExpr:
		fieldType = fmt.Sprintf("*%s", fi.X.(*ast.Ident).Name)
	case *ast.ArrayType:
		fieldType = fmt.Sprintf("[]%s", fieldTypeParser(fi.Elt))
	case *ast.MapType:
		fieldType = fmt.Sprintf("%s->%s", fieldTypeParser(fi.Key), fieldTypeParser(fi.Value))
	default:
		fmt.Printf("Handle this type: %T\n", fi)
	}

	return fieldType
}

func structTypeParser(strctype *ast.StructType) map[string]*DomainField {

	fields := make(map[string]*DomainField, strctype.Fields.NumFields())

	for _, field := range strctype.Fields.List {

		fieldType := fieldTypeParser(field.Type)
		var fieldName string

		for _, ident := range field.Names {
			if ident.Name != "" {
				fieldName = ident.Name
				break
			}
		}

		mapTypes := strings.Split(fieldType, "->")
		isMap := len(mapTypes) == 2

		if isMap {

			fields[fieldName] = &DomainField{
				Name:   fieldName,
				IsMap:  isMap,
				MapKey: getFieldType(mapTypes[0]),
				Type:   getFieldType(mapTypes[1]),
			}

		} else {

			fields[fieldName] = &DomainField{
				Name:  fieldName,
				Type:  getFieldType(mapTypes[0]),
				IsMap: false,
			}
		}

	}

	return fields
}

func getFieldType(fieldType string) FieldType {

	isSlicePointers := IsSliceOfPointers(fieldType)

	split := strings.Split(RemoveArray(RemovePointer(fieldType)), ".")
	pkg := ""
	ftype := split[0]
	if len(split) == 2 {
		pkg = split[0]
		ftype = split[1]
	}

	return FieldType{
		Package:       pkg,
		Type:          ftype,
		IsTypePointer: isSlicePointers || IsPointer(fieldType),
		IsSlice:       isSlicePointers || IsSlice(fieldType),
	}

}

func parseTree(n ast.Node, dtypes map[string]*DomainType) {

	switch x := n.(type) {
	case *ast.TypeSpec:
		switch types := x.Type.(type) {
		case *ast.StructType:
			if dtypes[x.Name.Name] == nil {
				break
			}
			dtypes[x.Name.Name].Fields = structTypeParser(types)
		default:
			// fmt.Println(x.Name.Name)
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

			k := RemovePointer(recv)
			dtype := dtypes[k]
			if dtype == nil {
				break
			}

			// funcName := fieldTypeParser(x.Name)
			// var resstr string

			// if x.Type.Results != nil {
			// 	for _, res := range x.Type.Results.List {
			// 		resstr = fieldTypeParser(res.Type)
			// 	}
			// }

			// isSlicePointers := IsSliceOfPointers(resstr)

			// split := strings.Split(RemoveArray(RemovePointer(resstr)), ".")
			// pkg := ""
			// ftype := split[0]
			// if len(split) == 2 {
			// 	pkg = split[0]
			// 	ftype = split[1]
			// }

			// getter := DomainFunc{
			// 	Recv: recv,
			// 	Name: funcName,
			// 	ResultType: FieldType{
			// 		Package:       pkg,
			// 		Type:          ftype,
			// 		IsTypePointer: isSlicePointers || IsPointer(resstr),
			// 		IsSlice:       isSlicePointers || IsSlice(resstr),
			// 	},
			// }

			// dtypes[k].Funcs = append(dtypes[k].Funcs, getter)
		}
	default:
		if x == nil {
			break
		}
		// fmt.Printf("%T\n", x)
	}

}
