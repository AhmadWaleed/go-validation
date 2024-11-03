package main

import (
	"bytes"
	"fmt"

	"github.com/spf13/cast"
)

type Value struct {
	Value any
}

func (v Value) toString() string {
	switch v.Value.(type) {
	case float32, float64:
		v := cast.ToString(v.Value)
		if v == "0.0" {
			return "0"
		}
		return v
	default:
		return cast.ToString(v.Value)
	}
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
	Cond1  *Value
	Cond2  *Value
}

type Generator struct {
	pkg *Package
	buf bytes.Buffer // Accumulated output.
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *Generator) generate(schema Schema) {
	// Generate _Gov_Rule interface and rule types with their Validate methods
	g.Printf("\n")
	g.Printf("type _Gov_Rule interface {\n")
	g.Printf("\tValidate() error\n")
	g.Printf("}\n")

	// Presence rule: no additional values
	g.Printf("\n")
	g.Printf("// presence          required                A rule without additional values\n")
	g.Printf("type _Gov_RulePresence struct {\n")
	g.Printf("\tField     string\n")
	g.Printf("\tValue     any\n")
	g.Printf("\tValidator _Gov_PresenceValidator\n")
	g.Printf("}\n")
	g.Printf("\n")
	g.Printf("func (r _Gov_RulePresence) Validate() error {\n")
	g.Printf("\treturn r.Validator(r.Field, r.Value)\n")
	g.Printf("}\n")

	// Value Constraint rule: single key-value pair (e.g., max:1000)
	g.Printf("\n")
	g.Printf("// value_constraint  max:1000                A rule with a single key-value pair\n")
	g.Printf("type _Gov_RuleValueConstraint struct {\n")
	g.Printf("\tName      string\n")
	g.Printf("\tField     string\n")
	g.Printf("\tValue     any\n")
	g.Printf("\tCond      any\n")
	g.Printf("\tValidator _Gov_ValueConstraintValidator\n")
	g.Printf("}\n")
	g.Printf("\n")
	g.Printf("func (r _Gov_RuleValueConstraint) Validate() error {\n")
	g.Printf("\treturn r.Validator(r.Field, r.Value, r.Cond)\n")
	g.Printf("}\n")

	// Range rule: specifies a range of values (e.g., between:1,1000)
	g.Printf("\n")
	g.Printf("// range             between:1,1000          A rule that specifies a range of values\n")
	g.Printf("type _Gov_RuleRange struct {\n")
	g.Printf("\tName      string\n")
	g.Printf("\tField     string\n")
	g.Printf("\tValue     any\n")
	g.Printf("\tMin       any\n")
	g.Printf("\tMax       any\n")
	g.Printf("\tValidator _Gov_RangeValidator\n")
	g.Printf("}\n")
	g.Printf("\n")
	g.Printf("func (r _Gov_RuleRange) Validate() error {\n")
	g.Printf("\treturn r.Validator(r.Field, r.Value, r.Min, r.Max)\n")
	g.Printf("}\n")

	// Conditional rule: depends on another field (e.g., required_if:Name=John)
	g.Printf("\n")
	g.Printf("// conditional       required_if:Name=John   A rule that depends on another field\n")
	g.Printf("type _Gov_RuleConditional struct {\n")
	g.Printf("\tName      string\n")
	g.Printf("\tField1    string\n")
	g.Printf("\tField2    string\n")
	g.Printf("\tValue1    any\n")
	g.Printf("\tValue2    any\n")
	g.Printf("\tCond      any\n")
	g.Printf("\tValidator _Gov_ConditionalValidator\n")
	g.Printf("}\n")
	g.Printf("\n")
	g.Printf("func (r _Gov_RuleConditional) Validate() error {\n")
	g.Printf("\treturn r.Validator(r.Field1, r.Value1, r.Field2, r.Value2, r.Cond)\n")
	g.Printf("}\n")

	// Generate type(s) _Gov_(*)Validator
	// Generate type _Gov_Rule
	// Generate type(s)  _Gov_Rule(*)
	// Generate _Gov_Schema_message map.
	// Generate _Gov_Error
	// Gennerate _Gov_(*) _Gov_(*)Validator
	// Generate type(s) (*)Schema & New(*)Schema(u (*))
}
