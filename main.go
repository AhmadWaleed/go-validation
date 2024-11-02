package main

import (
	"log"
)

func main() {
	u := User{ID: 0}
	schema := NewUserSchema(u)
	if err := schema.Validate(); err != nil {
		log.Fatalln(err)
	}
}

//go:generate govader --name User -output=schemas.go
type User struct {
	ID   int    `gov:"required"`
	Name string `gov:"required"`
	// ID    int    `gov:"required,min=1,max=1000,regexp=^[0-9]*$,required_if=Name=John,between=1-1000,different=ID2,size=10,same=ID3"`
	// Name  string `gov:"required,min=3,max=100,regexp=^[a-zA-Z]*$",required_with=ID",required_without=Email"`
	Email string `gov:"email"`
	// Age   int
}
