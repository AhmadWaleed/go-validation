// Generated by ego.
// DO NOT EDIT

//line tmpl.ego:1

package main

import "fmt"
import "html"
import "io"
import "context"

type Template struct {
	PackageName string
	Generator   *Generator
}

func (tmpl *Template) Render(ctx context.Context, w io.Writer) {
//line tmpl.ego:10
	_, _ = io.WriteString(w, "\n// Code generated by \"govader\"; DO NOT EDIT.\npackage ")
//line tmpl.ego:11
	_, _ = io.WriteString(w, html.EscapeString(fmt.Sprint(tmpl.PackageName)))
//line tmpl.ego:12
	_, _ = io.WriteString(w, "\n\nimport (\n\t\"errors\"\n\t\"regexp\"\n\t\"strings\"\n\n\t\"github.com/spf13/cast\"\n)\n\n// presence\t        required\t            A rule without additional values\n// value_constraint\tmax:1000\t            A rule with a single key-value pair\n// conditional\t    required_if:Name=John\tA rule that depends on another field\n// range\tbetween:1,1000\tA rule that specifies a range of values\ntype (\n\t_Gov_PresenceValidator[T any]\t\t\tfunc(field string, value T) error\n\t_Gov_ValueConstraintValidator[T any]\tfunc(field string, value T, cond T) error\n\t_Gov_RangeValidator[T any]           \tfunc(field string, value T, min T, max T) error\n\t_Gov_ConditionalValidator     \t\t\tfunc(field1 string, value1 any, field2 string, value2 any, cond any) error\n)\n\ntype _Gov_Rule interface {\n\tValidate() error\n}\n\n// presence\t        required\t            A rule without additional values\ntype _Gov_RulePresence [T any]struct {\n\tField     string\n\tValue     T\n\tValidator _Gov_PresenceValidator[T]\n}\n\nfunc (r _Gov_RulePresence[T]) Validate() error {\n\treturn r.Validator(r.Field, r.Value)\n}\n\n// value_constraint\tmax:1000\t            A rule with a single key-value pair\ntype _Gov_RuleValueConstraint [T any]struct {\n\tName      string\n\tField     string\n\tValue     T\n\tCond      T\n\tValidator _Gov_ValueConstraintValidator[T]\n}\n\nfunc (r _Gov_RuleValueConstraint[T]) Validate() error {\n\treturn r.Validator(r.Field, r.Value, r.Cond)\n}\n\n// range\tbetween:1,1000\tA rule that specifies a range of values\ntype _Gov_RuleRange[T any] struct {\n\tName      string\n\tField     string\n\tValue     T\n\tMin       T\n\tMax       T\n\tValidator _Gov_RangeValidator[T]\n}\n\nfunc (r _Gov_RuleRange[T]) Validate() error {\n\treturn r.Validator(r.Field, r.Value, r.Min, r.Max)\n}\n\n// conditional\t    required_if:Name=John\tA rule that depends on another field\ntype _Gov_RuleConditional struct {\n\tName      string\n\tField1    string\n\tField2    string\n\tValue1    any\n\tValue2    any\n\tCond      any\n\tValidator _Gov_ConditionalValidator\n}\n\nfunc (r _Gov_RuleConditional) Validate() error {\n\treturn r.Validator(r.Field1, r.Value1, r.Field2, r.Value2, r.Cond)\n}\n\n")
//line tmpl.ego:89
	tmpl.Generator.Generate()
//line tmpl.ego:90
	_, _ = io.WriteString(w, "\n\n")
//line tmpl.ego:91
}

var _ fmt.Stringer
var _ io.Reader
var _ context.Context
var _ = html.EscapeString