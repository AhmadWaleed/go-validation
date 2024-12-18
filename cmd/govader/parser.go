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
	Name      string      // Name of the struct.
	FieldList []FieldInfo // List of fields in the struct.
}

type FieldInfo struct {
	Name string          // Name of the field.
	Tag  string          // Validation tag. e.g `required;min=1`
	Type types.BasicInfo // Type of the field.
}

type Schema struct {
	Type       StructInfo
	Rules      []SchemaRule
	Validators []string
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

func (r SchemaRule) FuncName() string {
	if r.Type == ruleConditional {
		return fmt.Sprintf("_Gov_%s", r.Name)
	}
	return fmt.Sprintf("_Gov_%s_%s", r.Name, r.Cond1.TypeString())
}

var (
	// presetValConstRules contains list of predefined value constraint rules.
	presetValConstRules = []string{"email"}
)

func parseSchema(info []StructInfo) ([]Schema, error) {
	schemas := make([]Schema, 0, len(info))
	uniqRuleSet := make(map[string]struct{})

	for _, stct := range info {
		rules := make([]SchemaRule, 0, 10)
		for _, field := range stct.FieldList {
			ruleset := strings.Split(field.Tag, ";")
			for _, rulestr := range ruleset {
				rule, err := parseRule(field, rulestr)
				if err != nil {
					return nil, err
				}
				rules = append(rules, rule)
				uniqRuleSet[rule.Name] = struct{}{}
			}
		}
		schema := Schema{
			Type:       stct,
			Rules:      rules,
			Validators: make([]string, 0, len(uniqRuleSet)),
		}
		for k := range maps.Keys(uniqRuleSet) {
			schema.Validators = append(schema.Validators, k)
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func parseRule(f FieldInfo, rawRule string) (SchemaRule, error) {
	seprator := "=" // rule is either presence or value constraint or range.
	if strings.IndexRune(rawRule, ':') != -1 {
		seprator = ":" // rule is conditional.
	}

	kv := strings.SplitN(rawRule, seprator, 2)
	if len(kv) < 0 || len(kv) > 2 {
		return SchemaRule{}, fmt.Errorf("invalid rule format: %v", rawRule)
	}

	var rule SchemaRule
	if len(kv) == 1 /* Presense rule */ {
		rule = parsePresenceRule(f, kv[0])
	} else if strings.Contains(kv[1], ",") /* range */ {
		rule = parseRangeRule(f, kv[0], kv[1])
	} else if seprator == ":" /* Conditional rule */ {
		rule = parseConditionalRule(f, kv[0], kv[1])
	} else /* Value constraint */ {
		rule = parseValueConstraintRule(f, kv[0], kv[1])
	}

	return rule, nil
}

func parsePresenceRule(f FieldInfo, ruleName string) SchemaRule {
	if slices.Contains(presetValConstRules, ruleName) {
		return SchemaRule{
			Name:   ruleName,
			Type:   ruleValueConstraint,
			Field1: f.Name,
			Cond1:  &Value{Type: f.Type}, // We only need type to generate typed rule.
		}
	}
	return SchemaRule{
		Name:   ruleName,
		Type:   rulePresence,
		Field1: f.Name,
		Cond1:  &Value{Type: f.Type}, // We only need type to generate typed rule.
	}
}

func parseRangeRule(f FieldInfo, ruleName, ruleValue string) SchemaRule {
	min, max, _ := strings.Cut(ruleValue, ",")
	return SchemaRule{
		Name:   ruleName,
		Type:   ruleRange,
		Field1: f.Name,
		Cond1:  parseValue(f.Type, min),
		Cond2:  parseValue(f.Type, max),
	}
}

func parseConditionalRule(f FieldInfo, ruleName, ruleValue string) SchemaRule {
	if field2, cond, ok := strings.Cut(ruleValue, "="); ok {
		return SchemaRule{
			Name:   ruleName,
			Type:   ruleConditional,
			Field1: f.Name,
			Field2: field2,
			Cond1:  parseValue(f.Type, cond),
		}
	}
	return SchemaRule{
		Name:   ruleName,
		Type:   ruleConditional,
		Field1: f.Name,
		Field2: ruleValue,
	}
}

func parseValueConstraintRule(f FieldInfo, ruleName, ruleValue string) SchemaRule {
	return SchemaRule{
		Name:   ruleName,
		Type:   ruleValueConstraint,
		Field1: f.Name,
		Cond1:  parseValue(f.Type, ruleValue),
	}
}

func parseValue(t types.BasicInfo, v string) *Value {
	switch t {
	case types.IsString:
		if vv, err := cast.ToInt64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else {
			return &Value{Value: vv, Type: t}
		}
	case types.IsInteger:
		if vv, err := cast.ToInt64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else {
			return &Value{Value: vv, Type: t}
		}
	case types.IsUnsigned:
		if vv, err := cast.ToUint64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else {
			return &Value{Value: vv, Type: t}
		}
	case types.IsFloat:
		if vv, err := cast.ToFloat64E(v); err != nil {
			return &Value{Value: v, Type: t}
		} else {
			return &Value{Value: vv, Type: t}
		}
	default:
		return &Value{Value: v, Type: t}
	}
}
