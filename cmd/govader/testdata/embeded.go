package main

import (
	"reflect"
	"strings"
)

type Ident struct {
	ID string `gov:"required"`
}

type Embeded struct {
	Ident
	Flag uint `gov:"required"`
}

func main() {
	// Happy path, all rules passes.
	t1 := Embeded{
		Ident: Ident{ID: "1001"},
		Flag:  1,
	}
	schema1 := NewEmbededSchema(t1)
	ck(schema1.Validate(), []string(nil))

	// Fails all rules.
	t2 := Embeded{}
	schema2 := NewEmbededSchema(t2)
	ck(schema2.Validate(), []string{
		"The ID field is required.",
		"The Flag field is required.",
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
