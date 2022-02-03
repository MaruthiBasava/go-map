package main

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
)

func GenerateDomainMappers(domainTypes map[string]*DomainType) {

	// f := NewFile("mapping_gen")

	for _, domainType := range domainTypes {

		fmt.Println(domainType.Type, domainType.File)

		fields := make([]Code, len(domainType.Fields))
		i := 0

		mapFromDict := Dict{}
		mapToDict := Dict{}

		firstLetter := strings.ToLower(domainType.Type)[0:1]

		for _, field := range domainType.Fields {

			a := Id(field.Getter.Name)

			if field.Type.IsSlice {
				a = a.Index()
			}

			if field.Type.IsTypePointer {
				a = a.Op("*")
			}

			fields[i] = a.Qual(field.Type.Package, field.Type.Type)

			mapFromDict[Id(field.Name)] = Qual("output", field.Getter.Name)
			mapToDict[Id(field.Getter.Name)] = Id(firstLetter).Dot(field.Name)
			// fmt.Println("\t", field.Name, field.Type, field.Type.IsSlice, field.Type.IsTypePointer)
			// fmt.Println("\t\t", field.Getter)
			i++
		}

		strct := Type().Id(domainType.Type + "Output").Struct(fields...)
		fmt.Printf("%#v\n", strct)

		mapFrom := Func().Id("Map"+domainType.Type+"From").Params(
			Id("output").Qual("domain", domainType.Type+"Output"),
		).Op("*").Qual("domain", domainType.Type).Block(
			Id("d").Op(":=").Op("&").Qual("domain", domainType.Type).Values(mapFromDict),
			Return(Id("d")),
		)

		mapTo := Func().Params(
			Id(firstLetter).Op("*").Id(domainType.Type),
		).Id("MapToOutput").Params().Qual("domain", domainType.Type+"Output").Block(
			Id("d").Op(":=").Op("&").Qual("domain", domainType.Type+"Output").Values(mapToDict),
			Return(Id("d")),
		)

		fmt.Printf("%#v\n", mapFrom)
		fmt.Printf("%#v\n", mapTo)

	}

}
