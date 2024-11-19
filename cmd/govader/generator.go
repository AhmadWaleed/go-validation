package main

import (
	"bytes"
	"fmt"
)

type Generator struct {
	w              *bytes.Buffer // Accumulated output.
	Schema         Schema
	Messages       map[string]string
	GeneratedRules map[string]bool // To keep track of generated rules to avoid duplicates.
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(g.w, format, args...)
}

func (g *Generator) Generate() {
	// Generator schema locale messages.
	g.Printf("var _Gov_Schema_message = map[string]string{\n")
	for rule, msg := range g.Messages {
		g.Printf("\t\"%s\": \"%s\",\n", rule, msg)
	}
	g.Printf("}\n")

	// Generate error func to return rule error messages.
	g.Printf(`func _Gov_Error(key, field1, value1, field2, value2 string) error {
		var msg string
		for _, word := range strings.Split(_Gov_Schema_message[key], " ") {
			if !strings.HasPrefix(word, ":") {
				msg += word + " "
				continue
			}
			switch strings.Trim(word, ".") /* Remove trailing '.' */ {
			case ":field", ":field1":
				msg += field1 + " "
			case ":value", ":value1":
				msg += value1 + " "
			case ":field2":
				msg += field2 + " "
			case ":value2":
				msg += value2 + " "
			default:
			}
		}
		msg = strings.Trim(msg, " ")
		if !strings.HasSuffix(msg, ".") {
			msg = msg + "."
		}
		return errors.New(msg)
}`)
	g.Printf("\n\n")

	// Generate rules.
	for _, rule := range g.Schema.Rules {
		if _, ok := g.GeneratedRules[rule.Name]; !ok {
			g.GenRule(rule)
			g.Printf("\n")
			g.GeneratedRules[rule.Name] = true
		}
	}
}

func (g *Generator) GenRule(rule SchemaRule) {
	switch rule.Type {
	case rulePresence:
		g.GenPresenceRule(rule)
	case ruleValueConstraint:
		g.GenValueConstraintRule(rule)
	case ruleRange:
		g.GenRangeRule(rule)
	case ruleConditional:
		g.GenConditionalRule(rule)
	}
}

func (g *Generator) GenPresenceRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s) error {\n", rule.Name, typ, typ)
	switch typ {
	case "int64":
		g.Printf("\tif value == 0 {\n")
	case "unit64":
		g.Printf("\tif value <= 0 {\n")
	case "float64":
		g.Printf("\tif value == 0 || value == 0.0 {\n")
	case "string":
		g.Printf("\tif value == \"\" {\n")
	}
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenValueConstraintRule(rule SchemaRule) {
	switch rule.Name {
	case "min":
		g.GenMinRule(rule)
	case "max":
		g.GenMaxRule(rule)
	case "size":
		g.GenSizeRule(rule)
	case "regexp":
		g.GenRegexpRule(rule)
	case "email":
		g.GenEmailRule(rule)
	}
}

func (g *Generator) GenRangeRule(rule SchemaRule) {
	switch rule.Name {
	case "between":
		g.GenBetweenRule(rule)
	}
}

func (g *Generator) GenConditionalRule(rule SchemaRule) {
	switch rule.Name {
	case "required_if":
		g.GenRequiredIfRule(rule)
	case "required_with":
		g.GenRequiredWithRule(rule)
	case "required_without":
		g.GenRequiredWithoutRule(rule)
	case "same":
		g.GenSameRule(rule)
	case "different":
		g.GenDifferentRule(rule)
	}
}

func (g *Generator) GenRequiredIfRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, _ any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv2, c := cast.ToString(value2), cast.ToString(cond)\n")
	g.Printf("\tif v2 != c {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field1, \"\", field2, v2)\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenRequiredWithRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif v2 != \"\" {\n")
	g.Printf("\t\tif v1 == \"\" {\n")
	g.Printf("\t\t\treturn _Gov_Error(\"%s\", field1, \"\", field2, \"\")\n", rule.Name)
	g.Printf("\t\t}\n")
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenRequiredWithoutRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif v2 == \"\" {\n")
	g.Printf("\t\tif v1 == \"\" {\n")
	g.Printf("\t\t\treturn _Gov_Error(\"%s\", field1, v1, field2, v2)\n", rule.Name)
	g.Printf("\t\t}\n")
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenSameRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif value1 != value2 {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field1, v1, field2, v2)\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenDifferentRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif value1 == value2 {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field1, v1, field2, v2)\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenBetweenRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s) error {\n", rule.Name, typ, typ)
	switch typ {
	case "int64":
		g.Printf("\tif value < %d || value > %d {\n", rule.Cond1.Value, rule.Cond2.Value)
	case "unit64":
		g.Printf("\tif value <= %d || value >= %d {\n", rule.Cond1.Value, rule.Cond2.Value)
	case "float64":
		g.Printf("\tif value < %f || value > %f {\n", rule.Cond1.Value, rule.Cond2.Value)
	case "string":
		g.Printf("\tif len(value) < %d || len(value) > %d {\n", rule.Cond1.Value, rule.Cond2.Value)
	}
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenMinRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s) error {\n", rule.Name, typ, typ)
	switch typ {
	case "int64":
		g.Printf("\tif value < %d {\n", rule.Cond1.Value)
	case "unit64":
		g.Printf("\tif value <= %d {\n", rule.Cond1.Value)
	case "float64":
		g.Printf("\tif value < %f {\n", rule.Cond1.Value)
	case "string":
		g.Printf("\tif len(value) < %d {\n", rule.Cond1.Value)
	}
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenMaxRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s) error {\n", rule.Name, typ, typ)
	switch typ {
	case "int64":
		g.Printf("\tif value > %d {\n", rule.Cond1.Value)
	case "unit64":
		g.Printf("\tif value >= %d {\n", rule.Cond1.Value)
	case "float64":
		g.Printf("\tif value > %f {\n", rule.Cond1.Value)
	case "string":
		g.Printf("\tif len(value) > %d {\n", rule.Cond1.Value)
	}
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenSizeRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s) error {\n", rule.Name, typ, typ)
	switch typ {
	case "int64":
		g.Printf("\tif value != %d {\n", rule.Cond1.Value)
	case "unit64":
		g.Printf("\tif value != %d {\n", rule.Cond1.Value)
	case "float64":
		g.Printf("\tif value != %f {\n", rule.Cond1.Value)
	case "string":
		g.Printf("\tif len(value) != %d {\n", rule.Cond1.Value)
	}
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenRegexpRule(rule SchemaRule) {
	g.Printf("func _Gov_%s_string(field string, value string) error {\n", rule.Name)
	g.Printf("\tpattern := \"%s\"\n", rule.Cond1.Value)
	g.Printf("\tre := regexp.MustCompile(pattern)\n")
	g.Printf("\tif ok := re.MatchString(value); !ok {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, pattern, \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenEmailRule(rule SchemaRule) {
	g.Printf("func _Gov_%s_string(field string, value string) error {\n", rule.Name)
	g.Printf("\tatIndex, dotIndex := strings.Index(value, \"@\"), strings.LastIndex(value, \".\")\n")
	g.Printf("\tif atIndex < 1 || dotIndex < atIndex+2 || dotIndex+2 >= len(value) {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}
