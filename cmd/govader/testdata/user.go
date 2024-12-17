package main

import (
	"reflect"
	"strings"
)

type User struct {
	ID    int64  `gov:"required;min=1;max=1000;regexp=^[0-9]*$;between=1,1000;different:ID2;size=2"`
	ID2   int    `gov:"required_if:Name=John"`
	ID3   int    `gov:"required_with:ID"`
	ID4   int    `gov:"required_without:ID"`
	ID5   string `gov:"required_with:ID6"`
	ID6   string `gov:"-"`
	Name  string `gov:"required"`
	Email string `gov:"email"`
}

func main() {
	// Happy path, all rules passes.
	u0 := User{
		ID:    12,
		ID3:   22,
		Name:  "Jane",
		Email: "jane@gmail.com",
	}
	schema0 := NewUserSchema(u0)
	ck(schema0.Validate(), []string(nil))

	// Fails all rules.
	u1 := User{}
	schema1 := NewUserSchema(u1)
	ck(schema1.Validate(), []string{
		"The ID field is required.",
		"The ID field must be at least 1.",
		"The ID field must be between 1 and 1000.",
		"The ID field must be different from the ID2 field.",
		"The ID field must be of size 2.",
		"The ID4 field is required when ID is not present.",
		"The Name field is required.",
		"The Email field must be a valid email address.",
	})

	// Fails required_with rule.
	u2 := User{
		ID:    12,
		Name:  "Jane",
		Email: "jane@gmail.com",
	}
	schema2 := NewUserSchema(u2)
	ck(schema2.Validate(), []string{
		"The ID3 field is required when ID is present.",
	})

	// Fails required_with rule with non gov tagged field.
	u3 := User{
		ID:    12,
		ID3:   33,
		ID6:   "1001",
		Name:  "Jane",
		Email: "jane@gmail.com",
	}
	schema3 := NewUserSchema(u3)
	ck(schema3.Validate(), []string{
		"The ID5 field is required when ID6 is present.",
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
