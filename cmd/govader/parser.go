package main

import (
	"fmt"
	"go/types"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cast"
)

type StructInfo struct {
	name      string      // Name of the struct.
	fieldList []FieldInfo // List of fields in the struct.
}

type FieldInfo struct {
	name string          // Name of the field.
	tag  string          // Validation tag. e.g `required;min=1`
	typ  types.BasicInfo // Type of the field.
}

type Schema struct {
	Rules      []SchemaRule
	validators []string
}

type Value struct {
	Type  types.BasicInfo
	Value any
}

func (v Value) TypeString() string {
	switch v.Type {
	case types.IsInteger:
		return "int64"
	case types.IsString:
		return "string"
	case types.IsFloat:
		return "float64"
	case types.IsUnsigned:
		return "uint64"
	default:
		return "string"
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

var (
	// e.g: {RequiredIf: [Same, Required]}
	validatorRuleSet = map[string][]string{
		"required":         nil,
		"required_if":      {"required", "same"},
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
	// presetValConstRules contains list of predefined value constraint rules.
	presetValConstRules = []string{"email"}
)

func parseSchema(info []StructInfo) (Schema, error) {
	uniqRuleSet := make(map[string]struct{})
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
				uniqRuleSet[rule.Name] = struct{}{}
			}
		}
	}

	schema := Schema{
		Rules:      rules,
		validators: make([]string, 0, len(uniqRuleSet)),
	}
	for k := range maps.Keys(uniqRuleSet) {
		schema.validators = append(schema.validators, k)
	}
	return schema, nil
}

func parseRule(f FieldInfo, rawRule string) (SchemaRule, error) {
	seprator := "=" // rule is either presence or value constraint or range.
	if strings.IndexRune(rawRule, ':') != -1 {
		seprator = ":" // rule is conditional.
	}

	var rule SchemaRule
	kv := strings.SplitN(rawRule, seprator, 2)
	if len(kv) < 0 || len(kv) > 2 {
		return rule, fmt.Errorf("invalid rule format: %v", rawRule)
	}

	field1 := f.name
	if len(kv) == 1 /* Presense rule */ {
		if slices.Contains(presetValConstRules, kv[0]) {
			rule = SchemaRule{
				Name:   kv[0],
				Type:   ruleValueConstraint,
				Field1: field1,
				Cond1:  &Value{Type: f.typ}, // We only need type to generate typed rule.
			}
		} else {
			rule = SchemaRule{
				Name:   kv[0],
				Type:   rulePresence,
				Field1: field1,
				Cond1:  &Value{Type: f.typ}, // We only need type to generate typed rule.
			}
		}
	} else if len(kv) == 2 {
		if strings.Contains(kv[1], ",") /* range */ {
			if min, max, ok := strings.Cut(kv[1], ","); ok {
				rule = SchemaRule{
					Name:   kv[0],
					Type:   ruleRange,
					Field1: field1,
					Cond1:  parseValue(f.typ, min),
					Cond2:  parseValue(f.typ, max),
				}
			}
		} else if strings.Contains(kv[1], "=") /* Conditional rule */ {
			if field2, cond, ok := strings.Cut(kv[1], "="); ok {
				rule = SchemaRule{
					Name:   kv[0],
					Type:   ruleConditional,
					Field1: field1,
					Field2: field2,
					Cond1:  parseValue(f.typ, cond),
				}
			}
		} else /* Value constraint */ {
			if seprator == ":" {
				rule = SchemaRule{
					Name:   kv[0],
					Type:   ruleConditional,
					Field1: field1,
					Field2: kv[1],
				}
			} else if seprator == "=" {
				rule = SchemaRule{
					Name:   kv[0],
					Type:   ruleValueConstraint,
					Field1: field1,
					Cond1:  parseValue(f.typ, kv[1]),
				}
			}
		}
	}
	if rule.Name == "" {
		return rule, fmt.Errorf("invalid rule format: %v", rawRule)
	}
	return rule, nil
}

func parseValue(t types.BasicInfo, v string) *Value {
	switch t {
	case types.IsString:
		if vv, err := cast.ToInt64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else if err == nil {
			return &Value{Value: vv, Type: t}
		}
	case types.IsInteger:
		if vv, err := cast.ToInt64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else if err == nil {
			return &Value{Value: vv, Type: t}
		}
	case types.IsUnsigned:
		if vv, err := cast.ToUint64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else if err == nil {
			return &Value{Value: vv, Type: t}
		}
	case types.IsFloat:
		if vv, err := cast.ToFloat64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else if err == nil {
			return &Value{Value: vv, Type: t}
		}
	default:
		return &Value{Value: v, Type: t}
	}
	return nil
}
