package main

import (
	"bytes"
	"fmt"
)

type Generator struct {
	w              *bytes.Buffer // Accumulated output.
	Schemas        []Schema
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
	for _, schema := range g.Schemas {
		for _, rule := range schema.Rules {
			if _, ok := g.GeneratedRules[rule.FuncName()]; !ok {
				g.GenRule(rule)
				g.Printf("\n")
				g.GeneratedRules[rule.FuncName()] = true
			}
		}
	}

	for _, schema := range g.Schemas {
		g.GenSchmaValdation(schema)
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
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2, c := cast.ToString(value1), cast.ToString(value2), cast.ToString(cond)\n")
	g.Printf("\tif v2 == c {\n")
	g.Printf("\t\tif v1 == \"\" || v1 == \"0\" {\n")
	g.Printf("\t\t\treturn _Gov_Error(\"%s\", field1, \"\", field2, c)\n", rule.Name)
	g.Printf("\t\t}\n")
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenRequiredWithRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif v2 != \"\" && v2 != \"0\" {\n")
	g.Printf("\t\tif v1 == \"\" || v1 == \"0\" {\n")
	g.Printf("\t\t\treturn _Gov_Error(\"%s\", field1, \"\", field2, \"\")\n", rule.Name)
	g.Printf("\t\t}\n")
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenRequiredWithoutRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif v2 == \"\" || v2 == \"0\" {\n")
	g.Printf("\t\tif v1 == \"\" || v1 == \"0\" {\n")
	g.Printf("\t\t\treturn _Gov_Error(\"%s\", field1, v1, field2, v2)\n", rule.Name)
	g.Printf("\t\t}\n")
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenSameRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif v1 != v2 {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field1, v1, field2, v2)\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenDifferentRule(rule SchemaRule) {
	g.Printf("func _Gov_%s(field1 string, value1 any, field2 string, value2 any, cond any) error {\n", rule.Name)
	g.Printf("\tv1, v2 := cast.ToString(value1), cast.ToString(value2)\n")
	g.Printf("\tif v1 == v2 {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field1, v1, field2, v2)\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenBetweenRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value, min, max %s) error {\n", rule.Name, typ, typ)
	g.Printf("\tn, m := cast.ToString(min), cast.ToString(max)\n")
	switch typ {
	case "int64":
		g.Printf("\tif value < min || value > max {\n")
	case "unit64":
		g.Printf("\tif value <= min || value >= max {\n")
	case "float64":
		g.Printf("\tif value < min || value > max {\n")
	case "string":
		g.Printf("\tif len(value) < min || len(value) > max {\n")
	}
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", n, m)\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenMinRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s, cond %s) error {\n", rule.Name, typ, typ, typ)
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
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, cast.ToString(cond), \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenMaxRule(rule SchemaRule) {
	typ := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s, cond %s) error {\n", rule.Name, typ, typ, typ)
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
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, cast.ToString(cond), \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenSizeRule(rule SchemaRule) {
	t := rule.Cond1.TypeString()
	g.Printf("func _Gov_%s_%s(field string, value %s, cond %s) error {\n", rule.Name, t, t, t)
	g.Printf("\tv := cast.ToString(value)\n")
	g.Printf("\tif len(v) != int(cond) {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, cast.ToString(cond), \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenRegexpRule(rule SchemaRule) {
	g.Printf("func _Gov_%s_string(field string, value string, cond string) error {\n", rule.Name)
	g.Printf("\tpattern := \"%s\"\n", rule.Cond1.Value)
	g.Printf("\tre := regexp.MustCompile(pattern)\n")
	g.Printf("\tif ok := re.MatchString(value); !ok {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, pattern, \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenEmailRule(rule SchemaRule) {
	g.Printf("func _Gov_%s_string(field string, value string, cond %s) error {\n", rule.Name, rule.Cond1.TypeString())
	g.Printf("\tatIndex, dotIndex := strings.Index(value, \"@\"), strings.LastIndex(value, \".\")\n")
	g.Printf("\tif atIndex < 1 || dotIndex < atIndex+2 || dotIndex+2 >= len(value) {\n")
	g.Printf("\t\treturn _Gov_Error(\"%s\", field, \"\", \"\", \"\")\n", rule.Name)
	g.Printf("\t}\n")
	g.Printf("\treturn nil\n")
	g.Printf("}\n")
}

func (g *Generator) GenSchmaValdation(schema Schema) {
	// Define the schema struct type
	g.Printf("type %sSchema struct {\n", schema.Type.Name)
	g.Printf("\trules []_Gov_Rule\n")
	g.Printf("}\n\n")

	// Define the constructor function for the schema
	g.Printf("func New%sSchema(u %s) %sSchema {\n", schema.Type.Name, schema.Type.Name, schema.Type.Name)
	g.Printf("\treturn %sSchema{\n", schema.Type.Name)
	g.Printf("\t\trules: []_Gov_Rule{\n")

	for _, rule := range schema.Rules {
		switch rule.Type {
		case rulePresence:
			// Generate presence rule
			g.Printf("\t\t\t_Gov_RulePresence[%s]{\n", rule.Cond1.TypeString())
			g.Printf("\t\t\t\tField:     \"%s\",\n", rule.Field1)
			g.Printf("\t\t\t\tValue:     %s(u.%s),\n", rule.Cond1.TypeString(), rule.Field1)
			g.Printf("\t\t\t\tValidator: _Gov_required_%s,\n", rule.Cond1.TypeString())
			g.Printf("\t\t\t},\n")

		case ruleValueConstraint:
			// Generate value constraint rule (e.g., min, max, regexp)
			typ := rule.Cond1.TypeString()
			if rule.Name == "regexp" {
				typ = "string"
			}
			g.Printf("\t\t\t_Gov_RuleValueConstraint[%s]{\n", typ)
			g.Printf("\t\t\t\tField:     \"%s\",\n", rule.Field1)
			if rule.Name == "regexp" {
				g.Printf("\t\t\t\tValue:     cast.ToString(u.%s),\n", rule.Field1)
			} else {
				g.Printf("\t\t\t\tValue:     %s(u.%s),\n", typ, rule.Field1)
			}
			if rule.Cond1 != nil {
				// Handle condition only if not nil, handle non presetValConstRules.
				if rule.Name == "regexp" {
					g.Printf("\t\t\t\tCond:      `%s`,\n", rule.Cond1.Value)
				} else {
					if rule.Cond1.Value != nil {
						g.Printf("\t\t\t\tCond:      %v,\n", rule.Cond1.Value)
					}
				}
			}
			g.Printf("\t\t\t\tValidator: _Gov_%s_%s,\n", rule.Name, typ)
			g.Printf("\t\t\t},\n")

		case ruleRange:
			// Generate range rule (e.g., between)
			g.Printf("\t\t\t_Gov_RuleRange[%s]{\n", rule.Cond1.TypeString())
			g.Printf("\t\t\t\tField:     \"%s\",\n", rule.Field1)
			g.Printf("\t\t\t\tValue:     %s(u.%s),\n", rule.Cond1.TypeString(), rule.Field1)
			if rule.Cond1 != nil {
				g.Printf("\t\t\t\tMin:       %v,\n", rule.Cond1.Value)
			}
			if rule.Cond2 != nil {
				g.Printf("\t\t\t\tMax:       %v,\n", rule.Cond2.Value)
			}
			g.Printf("\t\t\t\tValidator: _Gov_between_%s,\n", rule.Cond1.TypeString())
			g.Printf("\t\t\t},\n")

		case ruleConditional:
			// Generate conditional rule
			g.Printf("\t\t\t_Gov_RuleConditional{\n")
			g.Printf("\t\t\t\tField1:    \"%s\",\n", rule.Field1)
			g.Printf("\t\t\t\tField2:    \"%s\",\n", rule.Field2)
			g.Printf("\t\t\t\tValue1:    u.%s,\n", rule.Field1)
			g.Printf("\t\t\t\tValue2:    u.%s,\n", rule.Field2)
			if rule.Cond1 != nil {
				g.Printf("\t\t\t\tCond:      \"%v\",\n", rule.Cond1.Value)
			}
			g.Printf("\t\t\t\tValidator: _Gov_%s,\n", rule.Name)
			g.Printf("\t\t\t},\n")
		}
	}

	// Close the rules slice and return statement
	g.Printf("\t\t},\n")
	g.Printf("\t}\n")
	g.Printf("}\n")
	g.Printf("\n")

	// Generate the Validate method for the schema.
	g.Printf("func (s %sSchema) Validate() (messages []string) {\n", schema.Type.Name)
	g.Printf("\tfor _, rule := range s.rules {\n")
	g.Printf("\t\tif err := rule.Validate(); err != nil {\n")
	g.Printf("\t\t\tmessages = append(messages, err.Error())\n")
	g.Printf("\t\t}\n")
	g.Printf("\t}\n")
	g.Printf("\treturn messages\n")
	g.Printf("}\n")
}
