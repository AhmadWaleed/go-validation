package main

//go:generate govader --name User --output=schemas.go --locale=en,ar
type User struct {
	ID    int    `gov:"required;min=1,max=1000;regexp=^[0-9]*$;required_if=Name=John,between=1,1000;different=ID2;size=10;same=ID3"`
	Name  string `gov:"required"`
	Email string `gov:"email"`
	// Age   int
}
