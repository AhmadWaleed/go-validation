package main

import (
	"bytes"
	"fmt"
)

type StructSchema struct {
	Name      string
	Rules     []SchemaRule
	FieldList map[string]Value
}

type Value struct {
	Value any
}

func (v Value) toString() string {
	return ""
}

type ruleType uint8

const (
	rulePresence = iota
	ruleValueConstraint
	ruleConditional
	ruleRange
)

type SchemaRule struct {
	Name   string
	Type   ruleType
	Field1 string
	Field2 string
	Cond   *Value
}

type Generator struct {
	pkg *Package
	buf bytes.Buffer // Accumulated output.
	// validators map[string][]string // e.g: {RequiredIf: [Same, Required]}
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) generate(schema Schema) {
	// Generate type(s) _Gov_(*)Validator
	// Generate type _Gov_Rule
	// Generate type(s)  _Gov_Rule(*)
	// Generate _Gov_Schema_message map.
	// Generate _Gov_Error
	// Gennerate _Gov_(*) _Gov_(*)Validator
	// Generate type(s) (*)Schema & New(*)Schema(u (*))
}
