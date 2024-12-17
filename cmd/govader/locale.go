package main

import (
	"embed"
	"encoding/json"
)

//go:embed locale.json
var localeFS embed.FS

func LoadLocale(locale string) map[string]string {
	m := make(map[string]map[string]string)
	data, err := localeFS.ReadFile("locale.json")
	if err != nil {
		panic((err))
	}
	if err := json.Unmarshal(data, &m); err != nil {
		panic(err)
	}
	return m[locale]
}
