package main

import (
	"fmt"
	"go/types"
	"maps"
	"strings"

	"github.com/spf13/cast"
)

type StructInfo struct {
	name      string
	fieldList []FieldInfo
}

type FieldInfo struct {
	name string
	tag  string
	typ  types.BasicInfo
}

type Schema struct {
	rules      []SchemaRule
	validators []string
}

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

// e.g: {RequiredIf: [Same, Required]}
var validators = map[string][]string{
	"required":         nil,
	"required_if":      []string{"required", "same"},
	"required_with":    nil,
	"required_without": nil,
	"min":              nil,
	"max":              nil,
	"between":          nil,
	"same":             nil,
	"different":        nil,
	"regexp":           nil,
	"email":            nil,
}

func parseSchema(info []StructInfo) (Schema, error) {
	validatorSet := make(map[string]struct{})
	rules := make([]SchemaRule, 0, 10)
	for _, stct := range info {
		for _, field := range stct.fieldList {
			ruleset := strings.Split(field.tag, ";")
			for _, rulestr := range ruleset {
				rule, err := parseRule(field, rulestr)
				if err != nil {
					return Schema{}, err
				}
				rules = append(rules, rule)
				validatorSet[rule.Name] = struct{}{}
			}
		}
	}

	schema := Schema{
		rules:      rules,
		validators: make([]string, len(validatorSet)),
	}
	for k := range maps.Keys(validatorSet) {
		schema.validators = append(schema.validators, k)
	}
	return schema, nil
}

func parseRule(f FieldInfo, rawRule string) (SchemaRule, error) {
	var rule SchemaRule
	kv := strings.SplitN(rawRule, "=", 1)
	if len(kv) < 0 || len(kv) > 2 {
		return rule, fmt.Errorf("invalid rule format: %v", rawRule)
	}
	field1 := f.name
	if len(kv) == 1 /* Presense rule */ {
		rule = SchemaRule{
			Name:   kv[0],
			Type:   rulePresence,
			Field1: field1,
		}
	} else if len(kv) == 2 {
		if strings.Contains(kv[1], ",") /* range */ {
			if min, max, ok := strings.Cut(kv[1], ","); ok {
				rule = SchemaRule{
					Name:   kv[0],
					Type:   ruleRange,
					Field1: field1,
					Cond1:  &Value{min},
					Cond2:  &Value{max},
				}
			}
		} else if strings.Contains(kv[1], "=") /* Conditional rule */ {
			if field2, cond, ok := strings.Cut(kv[1], "="); ok {
				rule = SchemaRule{
					Name:   kv[0],
					Type:   ruleConditional,
					Field1: field1,
					Field2: field2,
					Cond1:  &Value{cond},
				}
			}
		} else /* Value constraint */ {
			rule = SchemaRule{
				Name:   kv[0],
				Type:   ruleValueConstraint,
				Field1: field1,
				Cond1:  &Value{kv[1]},
			}
		}
	}
	return rule, nil
}
