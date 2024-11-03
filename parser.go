package main

import "go/types"

type StructInfo struct {
	name      string
	fieldList []FieldInfo
}

type FieldInfo struct {
	name string
	tag  string
	typ  types.BasicInfo
}

type Schema struct {
	types      StructInfo
	validators map[string][]string // e.g: {RequiredIf: [Same, Required]}
}

func ParseSchema(values []StructInfo) Schema {
	return Schema{}
}
