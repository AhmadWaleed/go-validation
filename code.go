package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

func NewUserSchema(u User) UserSchema {
	return UserSchema{
		rules: []_Gov_Rule{
			_Gov_RulePresence{
				Field:     "ID",
				Value:     1,
				Validator: _Gov_Required,
			},
			_Gov_RulePresence{
				Field:     "Name",
				Value:     "John",
				Validator: _Gov_Required,
			},
			_Gov_RuleValueConstraint{
				Field:     "Age",
				Value:     40,
				Cond:      80,
				Validator: _Gov_Min,
			},
			_Gov_RuleRange{
				Field:     "ID",
				Value:     10,
				Min:       1,
				Max:       5,
				Validator: _Gov_Between,
			},
			_Gov_RuleRange{
				Field:     "ID",
				Value:     10,
				Min:       1,
				Max:       5,
				Validator: _Gov_Between,
			},
			_Gov_RuleConditional{
				Field1:    "Name",
				Field2:    "Username",
				Value1:    "",
				Value2:    "Doe", // Actual
				Cond:      "Doe", // Expected
				Validator: _Gov_RequiredIf,
			},
			_Gov_RuleConditional{
				Field1:    "Name",
				Field2:    "Username",
				Value1:    "",
				Value2:    "Doe",
				Validator: _Gov_RequiredWith,
			},
			_Gov_RuleConditional{
				Field1:    "Name",
				Field2:    "Username",
				Value1:    "",
				Value2:    "",
				Validator: _Gov_RequiredWithout,
			},
			_Gov_RuleValueConstraint{
				Field:     "Email",
				Value:     "abc",
				Cond:      `^\S+@\S+\.\S+$`,
				Validator: _Gov_Regexp,
			},
			_Gov_RuleValueConstraint{
				Field:     "Email",
				Value:     "aaa",
				Validator: _Gov_Email,
			},
			_Gov_RuleConditional{
				Field1:    "Name",
				Field2:    "Username",
				Value1:    "a",
				Value2:    "a",
				Validator: _Gov_Different,
			},
			_Gov_RuleConditional{
				Field1:    "ID",
				Field2:    "Age",
				Value1:    1.0,
				Value2:    1.0,
				Validator: _Gov_Different,
			},
		},
	}

}

type UserSchema struct {
	User
	rules []_Gov_Rule
}

func (s UserSchema) Validate() (messages []string) {
	for _, rule := range s.rules {
		if err := rule.Validate(); err != nil {
			messages = append(messages, err.Error())
		}
	}
	return messages
}

// presence	        required	            A rule without additional values
// value_constraint	max:1000	            A rule with a single key-value pair
// conditional	    required_if:Name=John	A rule that depends on another field
// range	between:1,1000	A rule that specifies a range of values
type (
	_Gov_PresenceValidator        func(field string, value any) error
	_Gov_ValueConstraintValidator func(field string, value any, cond any) error
	_Gov_RangeValidator           func(field string, value any, min any, max any) error
	_Gov_ConditionalValidator     func(field1 string, value1 any, field2 string, value2 any, cond any) error
)

type _Gov_Rule interface {
	Validate() error
}

// presence	        required	            A rule without additional values
type _Gov_RulePresence struct {
	Field     string
	Value     any
	Validator _Gov_PresenceValidator
}

func (r _Gov_RulePresence) Validate() error {
	return r.Validator(r.Field, r.Value)
}

// value_constraint	max:1000	            A rule with a single key-value pair
type _Gov_RuleValueConstraint struct {
	Name      string
	Field     string
	Value     any
	Cond      any
	Validator _Gov_ValueConstraintValidator
}

func (r _Gov_RuleValueConstraint) Validate() error {
	return r.Validator(r.Field, r.Value, r.Cond)
}

// range	between:1,1000	A rule that specifies a range of values
type _Gov_RuleRange struct {
	Name      string
	Field     string
	Value     any
	Min       any
	Max       any
	Validator _Gov_RangeValidator
}

func (r _Gov_RuleRange) Validate() error {
	return r.Validator(r.Field, r.Value, r.Min, r.Max)
}

// conditional	    required_if:Name=John	A rule that depends on another field
type _Gov_RuleConditional struct {
	Name      string
	Field1    string
	Field2    string
	Value1    any
	Value2    any
	Cond      any
	Validator _Gov_ConditionalValidator
}

func (r _Gov_RuleConditional) Validate() error {
	return r.Validator(r.Field1, r.Value1, r.Field2, r.Value2, r.Cond)
}

var _Gov_Schema_message = map[string]string{
	"required":         "The :field field is required.",
	"required_if":      "The :field1 field is required when :field2 is :value2.",
	"required_with":    "The :field1 field is required when :field2 is present.",
	"required_without": "The :field1 field is required when :field2 is not present.",
	"min":              "The :field field must be at least :value.",
	"max":              "The :field field may not be greater than :value.",
	"size":             "The :field field must be :value.",
	"same":             "The :field1 field must match the :field2 field.",
	"different":        "The :field1 field must be different from the :field2 field.",
	"between":          "The :field1 field must be between :field2 and :value2.",
	"regexp":           "The :field field does not match the required format :value.",
	"email":            "The :field field must be a valid email address.",
}

// _Gov_Error returns an error message based on the given key.
func _Gov_Error(key, field1, value1, field2, value2 string) error {
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
}

func _Gov_Required(field string, value any) error {
	err := _Gov_Error("required", field, "", "", "")
	switch value.(type) {
	case int, int32, int64, float32, float64:
		// float64 is used as Canonical form.
		if v := cast.ToFloat64(value); v == 0 {
			return err
		}
	case uint, uint8, uint16, uint32, uint64:
		if v := cast.ToFloat64(value); v <= 0 {
			return err
		}
	case string:
		if v, ok := value.(string); ok && v == "" {
			return err
		}
	case nil:
		return err
	}
	return nil
}
func _Gov_Min(field string, value, cond any) error {
	v, c := cast.ToFloat64(value), cast.ToFloat64(cond)
	if v < c {
		return _Gov_Error("min", field, cast.ToString(cond), "", "")
	}
	return nil
}
func _Gov_Max(field string, value, cond any) error {
	v, c := cast.ToFloat64(value), cast.ToFloat64(cond)
	if v > c {
		return _Gov_Error("max", field, cast.ToString(cond), "", "")
	}
	return nil
}
func _Gov_Between(field string, value, min, max any) error {
	v, n, m := cast.ToFloat64(value), cast.ToFloat64(min), cast.ToFloat64(max)
	if v < n || v > m {
		return _Gov_Error("between", field, cast.ToString(v), cast.ToString(min), cast.ToString(max))
	}
	return nil
}
func _Gov_RequiredIf(field1 string, value1 any, field2 string, value2 any, cond any) error {
	v1, c := cast.ToString(value2), cast.ToString(cond)
	if err := _Gov_Same(field1, value2, field2, c, nil); err == nil {
		if err := _Gov_Required(field1, value1); err != nil {
			return _Gov_Error("required_if", field1, v1, field2, c)
		}
	}
	return nil
}
func _Gov_RequiredWith(field1 string, value1 any, field2 string, value2 any, cond any) error {
	v1, v2 := cast.ToString(value1), cast.ToString(value2)
	if v2 != "" {
		if v1 == "" {
			return _Gov_Error("required_with", field1, "", field2, "")
		}
	}
	return nil
}
func _Gov_RequiredWithout(field1 string, value1 any, field2 string, value2 any, cond any) error {
	v1, v2 := cast.ToString(value1), cast.ToString(value2)
	if v2 == "" {
		if v1 == "" {
			return _Gov_Error("required_without", field1, v1, field2, v2)
		}
	}
	return nil
}
func _Gov_Same(field1 string, value1 any, field2 string, value2 any, _ any) error {
	v1, v2 := cast.ToString(value1), cast.ToString(value2)
	if value1 != value2 {
		return _Gov_Error("same", field1, v1, field2, v2)
	}
	return nil
}
func _Gov_Regexp(field string, value any, cond any) error {
	pattern := cast.ToString(cond)
	re := regexp.MustCompile(pattern)
	if ok := re.MatchString(cast.ToString(value)); !ok {
		return _Gov_Error("regexp", field, pattern, "", "")
	}
	return nil
}
func _Gov_Email(field string, value any, _ any) error {
	v := cast.ToString(value)
	atIndex, dotIndex := strings.Index(v, "@"), strings.LastIndex(v, ".")
	if atIndex < 1 || dotIndex < atIndex+2 || dotIndex+2 >= len(v) {
		return _Gov_Error("email", field, "", "", "")
	}
	return nil
}
func _Gov_Different(field1 string, value1 any, field2 string, value2 any, _ any) error {
	v1, v2 := cast.ToString(value1), cast.ToString(value2)
	if v1 == v2 {
		return _Gov_Error("different", field1, v1, field2, v2)
	}
	return nil
}
