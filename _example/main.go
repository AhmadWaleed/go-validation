package main

import "log"

func main() {
	u := User{}
	schema := NewUserSchema(u)
	if errors := schema.Validate(); errors != nil {
		log.Println(errors)
	}
}

//go:generate govader -type=User -locale=en
type User struct {
	ID    int64 `gov:"required;min=1;max=1000;regexp=^[0-9]*$;required_if:Name=John;between=1,1000;different:ID2;size=10;same:ID3"`
	ID2   int
	ID3   int
	Name  string `gov:"required"`
	Email string `gov:"email"`
}
