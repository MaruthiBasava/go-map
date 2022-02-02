package main

type DepField struct {
	Name string
	Type string
	Tags map[string]string
}

type DepType struct {
	Type          string
	ImportFromDTO bool
	MapToDomain   bool
	MapFromDomain bool
	Fields        []DepField
}

type DepStructTag struct {
	Type        string
	IsSnakeCase bool
}

type DepConfig struct {
	Dir        string
	Package    string
	Filename   string
	StructTags []DepStructTag
	Types      []DepType
}
