package main

import (
	"embed"
	"encoding/json"
	"slices"
)

//go:embed locale.json
var locale embed.FS

func LoadLocale(langs ...string) map[string]map[string]string {
	m := make(map[string]map[string]string)
	data, err := locale.ReadFile("locale.json")
	if err != nil {
		panic((err))
	}
	if err := json.Unmarshal(data, &m); err != nil {
		panic(err)
	}
	for lang := range m {
		if !slices.Contains(langs, lang) {
			delete(m, lang)
		}
	}
	return m
}
