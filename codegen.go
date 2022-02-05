package main

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
)

type NonAggregateDTO struct {
	name    string
	mapFrom map[string]string
	mapTo   map[string]string
}

func (a NonAggregateDTO) MapFromToDict(receiver *Statement) Dict {

	dict := Dict{}

	for k, v := range a.mapFrom {
		r := *receiver
		dict[Id(k)] = (&r).Dot(v)
	}

	return dict
}

func (a NonAggregateDTO) MapToDict(receiver *Statement) Dict {

	dict := Dict{}

	for k, v := range a.mapTo {
		r := *receiver
		dict[Id(k)] = (&r).Dot(v)
	}

	return dict
}

func GenerateDomainMappers(dc *DomainConfig) {

	nonaggres := make(map[string]NonAggregateDTO)

	for name, dtoType := range dc.DTOTypes {

		nonaggres[name] = NonAggregateDTO{
			name:    name,
			mapFrom: make(map[string]string),
			mapTo:   make(map[string]string),
		}

		for _, field := range dtoType.Fields {
			nonaggres[name].mapFrom[field.MappingTo] = field.Name
			nonaggres[name].mapTo[field.Name] = field.MappingTo
		}

	}

	for _, dtoType := range dc.DTOTypes {

		pkg := dc.Package

		dtoName := dtoType.Type + dc.DTOSuffix

		fields := make([]Code, len(dtoType.Fields))
		i := 0

		mapFromDict := Dict{}
		mapToDict := Dict{}

		firstLetter := strings.ToLower(dtoType.Type)[0:1]

		mapFromInnerMapping := make([]Code, 0)
		mapToInnerMapping := make([]Code, 0)

		getters := make([]Code, 0)

		dtype := dc.DomainTypes[dtoType.Type]

		if dtype != nil {
			for _, dfield := range dc.DomainTypes[dtoType.Type].Fields {

				rettype := &Statement{}

				if dfield.IsMap {

					// mapBinding := dtoType.MapBindings[field.Name]
					rettype = Map(Qual(dfield.MapKey.Package, dfield.MapKey.Type))

				}

				if dfield.Type.IsSlice {
					rettype = rettype.Index()
				}

				if dfield.Type.IsTypePointer {
					rettype = rettype.Op("*")
				}

				if rettype != nil {
					rettype = rettype.Qual(dfield.Type.Package, dfield.Type.Type)
				} else {
					rettype = Qual(dfield.Type.Package, dfield.Type.Type)
				}

				get := Func().Params(
					Id(firstLetter).Op("*").Id(dtype.Type),
				).Id(strings.Title(dfield.Name)).Params().Add(rettype).Block(
					Return(Id(firstLetter).Dot(dfield.Name)),
				)

				getters = append(getters, get)
			}
		}
		for _, field := range dtoType.Fields {

			if field.Name == "" {
				continue
			}

			a := Id(field.Name)

			mapFromVal := Qual("o", field.Name)
			mapToVal := Qual(firstLetter, field.MappingTo)
			name := LowercaseFirstLetter(field.Name)

			if field.IsMap {

				// mapBinding := dtoType.MapBindings[field.Name]
				a = a.Map(Qual(field.MapKey.Package, field.MapKey.Type))

				if field.Type.IsSlice {
					a = a.Index()
				}

				if field.Type.IsTypePointer {
					a = a.Op("*")
				}

				mapType := Map(Qual(field.MapKey.Package, field.MapKey.Type))
				if dtoType.Func == "" {
					mapType.Qual(field.Type.Package, field.Type.Type+dc.DTOSuffix)
				} else {
					mapType.Qual(field.Type.Package, field.Type.Type)
				}

				// mFrom := Id(name).Op(":=").Make(mapType)
				// mFromMapping := For(
				// 	Id("key").Op(":=").Range().Qual("o", field.Name),
				// ).Block(

				// )

			} else if field.Type.IsSlice {
				a = a.Index()

				nonagg := nonaggres[field.Type.Type]
				mapFromVal = Qual("", name)
				mapToVal = Qual("", name)
				// shortName := name[0:3]
				mFrom := Id(name).Op(":=").Make(
					Index().Op("*").Qual(pkg, field.Type.Type).Op(",").Len(
						Qual("o", field.Name),
					),
				)

				mMapping := For(
					Id("i").Op(":=").Lit(0),
					Id("i").Op("<").Len(Qual("o", field.Name)),
					Id("i").Op("++"),
					// Id("index").Op(",").Id(shortName).Op(":=").Range().Qual("o", field.Name),
				).Block(
					Id(name).Index(Id("i")).Op("=").Op("&").Qual(pkg, field.Type.Type).Values(
						nonagg.MapFromToDict(
							Qual("o", field.Name).Index(Id("i")),
						),
					),
					// Id(name).Index(Id("index")).Op("=").Id(shortName[0:1]),
				)

				mapFromInnerMapping = append(mapFromInnerMapping, mFrom)
				mapFromInnerMapping = append(mapFromInnerMapping, mMapping)

				mTo := Id(name).Op(":=").Make(
					Index().Qual(pkg, field.Type.Type+dc.DTOSuffix).Op(",").Len(
						Qual(firstLetter, field.MappingTo),
					),
				)

				mToMapping := For(
					Id("i").Op(":=").Lit(0),
					Id("i").Op("<").Len(Qual(firstLetter, field.Name)),
					Id("i").Op("++"),
				).Block(
					Id(name).Index(Id("i")).Op("=").Qual(pkg, field.Type.Type+dc.DTOSuffix).Values(
						nonagg.MapToDict(
							Qual(firstLetter, field.MappingTo).Index(Id("i")),
						),
					),
				)

				mapToInnerMapping = append(mapToInnerMapping, mTo)
				mapToInnerMapping = append(mapToInnerMapping, mToMapping)

			}

			// if field.Type.IsTypePointer {
			// 	a = a.Op("*")
			// }

			if dt := dc.DTOTypes[field.Type.Type]; dt != nil && dt.Func == "" {
				fields[i] = a.Qual(field.Type.Package, dt.Type+dc.DTOSuffix)
			} else {
				// fmt.Println(field.Type.Package)
				fields[i] = a.Qual(field.Type.Package, field.Type.Type)
			}

			mapFromDict[Id(field.MappingTo)] = mapFromVal
			mapToDict[Id(field.Name)] = mapToVal
			// fmt.Println("\t", field.Name, field.Type, field.Type.IsSlice, field.Type.IsTypePointer)
			// fmt.Println("\t\t", field.Getter)
			i++
		}

		mapDomainFrom := strings.Replace(dc.MapFromFunc, "{domain_type}", dtoType.Type, 1)
		mapToOutput := strings.Replace(dc.MapToFunc, "{suffix}", dc.DTOSuffix, 1)

		if dtoType.Func == "" {
			strct := Type().Id(dtoName).Struct(fields...)
			fmt.Printf("%#v\n", strct)

			for _, g := range getters {
				fmt.Printf("%#v\n", g)
			}

		}

		if !dtoType.IsAggregateRoot {
			continue
		}

		structMappingFrom := Id("d").Op(":=").Op("&").Qual(pkg, dtoType.Type).Values(mapFromDict)
		mapFromInnerMapping = append(mapFromInnerMapping, structMappingFrom)

		returnCode := Return(Id("d"))
		mapFromInnerMapping = append(mapFromInnerMapping, returnCode)

		structMappingTo := Id("d").Op(":=").Op("&").Qual(pkg, dtoName).Values(mapToDict)
		mapToInnerMapping = append(mapToInnerMapping, structMappingTo)
		mapToInnerMapping = append(mapToInnerMapping, returnCode)

		mapFrom := Func().Id(mapDomainFrom).Params(
			Id("output").Qual(pkg, dtoName),
		).Op("*").Qual(pkg, dtoType.Type).Block(
			mapFromInnerMapping...,
		)

		mapTo := Func().Params(
			Id(firstLetter).Op("*").Id(dtoType.Type),
		).Id(mapToOutput).Params().Qual(pkg, dtoName).Block(
			mapToInnerMapping...,
		)

		fmt.Printf("%#v\n", mapFrom)
		fmt.Printf("%#v\n", mapTo)

	}

}
