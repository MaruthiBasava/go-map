package main

import (
	"fmt"
	"io/ioutil"
	"path"
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
		dict[Id(k)] = receiver.Clone().Dot(v)
	}

	return dict
}

func (a NonAggregateDTO) MapToDict(receiver *Statement) Dict {

	dict := Dict{}

	for k, v := range a.mapTo {
		dict[Id(k)] = receiver.Clone().Dot(v)
	}

	return dict
}

func GenerateDomainMappers(dc *DomainConfig) {

	fp := path.Join(dc.Dir, dc.Filename)
	f := NewFile(dc.Package)

	fmt.Println("*", dc.Imports)

	for imp, url := range dc.Imports {
		f.ImportName(url, imp)
	}

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

				fmt.Println("**", dfield.Name, dfield.Type.Package, dfield.Type.Type)

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

			fieldPkg := field.Type.Package

			if fieldPkg == pkg {
				fieldPkg = ""
			}

			outputId := "output"

			mapFromVal := Id(outputId).Dot(field.Name)
			mapToVal := Id(firstLetter).Dot(field.MappingTo)
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
					mapType.Qual(fieldPkg, field.Type.Type+dc.DTOSuffix)
				} else {
					mapType.Qual(fieldPkg, field.Type.Type)
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
					Index().Op("*").Id(field.Type.Type).Op(",").Len(
						Id(outputId).Dot(field.Name),
					),
				)

				mMapping := For(
					Id("i").Op(":=").Lit(0),
					Id("i").Op("<").Len(Id(outputId).Dot(field.Name)),
					Id("i").Op("++"),
					// Id("index").Op(",").Id(shortName).Op(":=").Range().Qual("o", field.Name),
				).Block(
					Id(name).Index(Id("i")).Op("=").Op("&").Id(field.Type.Type).Values(
						nonagg.MapFromToDict(
							Id(outputId).Dot(field.Name).Index(Id("i")),
						),
					),
					// Id(name).Index(Id("index")).Op("=").Id(shortName[0:1]),
				)

				mapFromInnerMapping = append(mapFromInnerMapping, mFrom)
				mapFromInnerMapping = append(mapFromInnerMapping, mMapping)

				mTo := Id(name).Op(":=").Make(
					Index().Id(field.Type.Type + dc.DTOSuffix).Op(",").Len(
						Id(firstLetter).Dot(field.MappingTo),
					),
				)

				mToMapping := For(
					Id("i").Op(":=").Lit(0),
					Id("i").Op("<").Len(Id(firstLetter).Dot(field.Name).Call()),
					Id("i").Op("++"),
				).Block(
					Id(name).Index(Id("i")).Op("=").Id(field.Type.Type + dc.DTOSuffix).Values(
						nonagg.MapToDict(
							Id(firstLetter).Dot(field.MappingTo).Index(Id("i")),
						),
					),
				)

				fmt.Printf("%#v\n", nonagg.MapToDict(
					Id(firstLetter).Dot(field.MappingTo).Index(Id("i")),
				))

				mapToInnerMapping = append(mapToInnerMapping, mTo)
				mapToInnerMapping = append(mapToInnerMapping, mToMapping)

			}

			if dt := dc.DTOTypes[field.Type.Type]; dt != nil && dt.Func == "" {
				fields[i] = a.Qual(fieldPkg, dt.Type+dc.DTOSuffix)
			} else {
				// fmt.Println(fieldPkg)
				fields[i] = a.Qual(fieldPkg, field.Type.Type)
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
			f.Type().Id(dtoName).Struct(fields...)
		}

		for _, g := range getters {
			f.Add(g)
		}

		if !dtoType.IsAggregateRoot {
			continue
		}

		nName := firstLetter + dc.DTOSuffix[0:1]

		structMappingFrom := Id(firstLetter).Op(":=").Op("&").Id(dtoType.Type).Values(mapFromDict)
		mapFromInnerMapping = append(mapFromInnerMapping, structMappingFrom)
		mapFromInnerMapping = append(mapFromInnerMapping, Return(Id(firstLetter)))

		structMappingTo := Id(nName).Op(":=").Id(dtoName).Values(mapToDict)
		mapToInnerMapping = append(mapToInnerMapping, structMappingTo)
		mapToInnerMapping = append(mapToInnerMapping, Return(Id(nName)))

		f.Func().Id(mapDomainFrom).Params(
			Id("output").Id(dtoName),
		).Op("*").Id(dtoType.Type).Block(
			mapFromInnerMapping...,
		)

		f.Func().Params(
			Id(firstLetter).Op("*").Id(dtoType.Type),
		).Id(mapToOutput).Params().Id(dtoName).Block(
			mapToInnerMapping...,
		)

	}

	fmt.Printf("%#v\n", f)

	ioutil.WriteFile(fp, []byte(fmt.Sprintf("%#v\n", f)), 0644)

}
