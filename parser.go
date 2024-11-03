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

func ParseSchema(info []StructInfo) []StructSchema {
	return nil
}
