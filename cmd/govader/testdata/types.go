package main

import (
	"reflect"
	"strings"
)

type Types struct {
	Int     int     `gov:"required"`
	Int8    int8    `gov:"required"`
	Int16   int16   `gov:"required"`
	Int32   int32   `gov:"required"`
	Int64   int64   `gov:"required"`
	Uint8   uint8   `gov:"required"`
	Uint16  uint16  `gov:"required"`
	Uint32  uint32  `gov:"required"`
	Uint64  uint64  `gov:"required"`
	Float32 float32 `gov:"required"`
	Float64 float64 `gov:"required"`
	String  string  `gov:"required"`
	// Boolean bool    `gov:"required"`
}

func main() {
	t1 := Types{}
	schema1 := NewTypesSchema(t1)
	ck(schema1.Validate(), []string{
		"The Int field is required.",
		"The Int8 field is required.",
		"The Int16 field is required.",
		"The Int32 field is required.",
		"The Int64 field is required.",
		"The Uint8 field is required.",
		"The Uint16 field is required.",
		"The Uint32 field is required.",
		"The Uint64 field is required.",
		"The Float32 field is required.",
		"The Float64 field is required.",
		"The String field is required.",
	})
}

func ck(got, want []string) {
	if ok := reflect.DeepEqual(want, got); !ok {
		panic(
			"user.go:\n" +
				"want:\n" + strings.Join(want, "\n") +
				"\n" +
				"got:\n" + strings.Join(got, "\n"),
		)
	}
}
