// Code generated by "govader"; DO NOT EDIT.
package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

// presence	        required	            A rule without additional values
// value_constraint	max:1000	            A rule with a single key-value pair
// conditional	    required_if:Name=John	A rule that depends on another field
// range	between:1,1000	A rule that specifies a range of values
type (
	_Gov_PresenceValidator[T any]        func(field string, value T) error
	_Gov_ValueConstraintValidator[T any] func(field string, value T, cond T) error
	_Gov_RangeValidator[T any]           func(field string, value T, min T, max T) error
	_Gov_ConditionalValidator            func(field1 string, value1 any, field2 string, value2 any, cond any) error
)

type _Gov_Rule interface {
	Validate() error
}

// presence	        required	            A rule without additional values
type _Gov_RulePresence[T any] struct {
	Field     string
	Value     T
	Validator _Gov_PresenceValidator[T]
}

func (r _Gov_RulePresence[T]) Validate() error {
	return r.Validator(r.Field, r.Value)
}

// value_constraint	max:1000	            A rule with a single key-value pair
type _Gov_RuleValueConstraint[T any] struct {
	Name      string
	Field     string
	Value     T
	Cond      T
	Validator _Gov_ValueConstraintValidator[T]
}

func (r _Gov_RuleValueConstraint[T]) Validate() error {
	return r.Validator(r.Field, r.Value, r.Cond)
}

// range	between:1,1000	A rule that specifies a range of values
type _Gov_RuleRange[T any] struct {
	Name      string
	Field     string
	Value     T
	Min       T
	Max       T
	Validator _Gov_RangeValidator[T]
}

func (r _Gov_RuleRange[T]) Validate() error {
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
	"min":              "The :field field must be at least :value.",
	"max":              "The :field field may not be greater than :value.",
	"size":             "The :field field must be of size :value.",
	"same":             "The :field1 field must match the :field2 field.",
	"different":        "The :field1 field must be different from the :field2 field.",
	"regexp":           "The :field field does not match the required format :value.",
	"required":         "The :field field is required.",
	"required_if":      "The :field1 field is required when :field2 is :value2.",
	"required_with":    "The :field1 field is required when :field2 is present.",
	"required_without": "The :field1 field is required when :field2 is not present.",
	"between":          "The :field1 field must be between :field2 and :value2.",
	"email":            "The :field field must be a valid email address.",
}

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

func _Gov_required_int64(field string, value int64) error {
	if value == 0 {
		return _Gov_Error("required", field, "", "", "")
	}
	return nil
}

func _Gov_min_int64(field string, value int64, cond int64) error {
	if value < 1 {
		return _Gov_Error("min", field, cast.ToString(cond), "", "")
	}
	return nil
}

func _Gov_max_int64(field string, value int64, cond int64) error {
	if value > 1000 {
		return _Gov_Error("max", field, cast.ToString(cond), "", "")
	}
	return nil
}

func _Gov_regexp_string(field string, value string, cond string) error {
	pattern := "^[0-9]*$"
	re := regexp.MustCompile(pattern)
	if ok := re.MatchString(value); !ok {
		return _Gov_Error("regexp", field, pattern, "", "")
	}
	return nil
}

func _Gov_required_if(field1 string, value1 any, field2 string, value2 any, cond any) error {
	v1, v2, c := cast.ToString(value1), cast.ToString(value2), cast.ToString(cond)
	if v2 == c {
		if v1 == "" || v1 == "0" {
			return _Gov_Error("required_if", field1, "", field2, c)
		}
	}
	return nil
}

func _Gov_between_int64(field string, value, min, max int64) error {
	n, m := cast.ToString(min), cast.ToString(max)
	if value < min || value > max {
		return _Gov_Error("between", field, "", n, m)
	}
	return nil
}

func _Gov_different(field1 string, value1 any, field2 string, value2 any, cond any) error {
	v1, v2 := cast.ToString(value1), cast.ToString(value2)
	if v1 == v2 {
		return _Gov_Error("different", field1, v1, field2, v2)
	}
	return nil
}

func _Gov_size_int64(field string, value int64, cond int64) error {
	v := cast.ToString(value)
	if len(v) != int(cond) {
		return _Gov_Error("size", field, cast.ToString(cond), "", "")
	}
	return nil
}

func _Gov_same(field1 string, value1 any, field2 string, value2 any, cond any) error {
	v1, v2 := cast.ToString(value1), cast.ToString(value2)
	if v1 != v2 {
		return _Gov_Error("same", field1, v1, field2, v2)
	}
	return nil
}

func _Gov_required_string(field string, value string) error {
	if value == "" {
		return _Gov_Error("required", field, "", "", "")
	}
	return nil
}

func _Gov_email_string(field string, value string, cond string) error {
	atIndex, dotIndex := strings.Index(value, "@"), strings.LastIndex(value, ".")
	if atIndex < 1 || dotIndex < atIndex+2 || dotIndex+2 >= len(value) {
		return _Gov_Error("email", field, "", "", "")
	}
	return nil
}

type UserSchema struct {
	rules []_Gov_Rule
}

func NewUserSchema(u User) UserSchema {
	return UserSchema{
		rules: []_Gov_Rule{
			_Gov_RulePresence[int64]{
				Field:     "ID",
				Value:     int64(u.ID),
				Validator: _Gov_required_int64,
			},
			_Gov_RuleValueConstraint[int64]{
				Field:     "ID",
				Value:     int64(u.ID),
				Cond:      1,
				Validator: _Gov_min_int64,
			},
			_Gov_RuleValueConstraint[int64]{
				Field:     "ID",
				Value:     int64(u.ID),
				Cond:      1000,
				Validator: _Gov_max_int64,
			},
			_Gov_RuleValueConstraint[string]{
				Field:     "ID",
				Value:     cast.ToString(u.ID),
				Cond:      `^[0-9]*$`,
				Validator: _Gov_regexp_string,
			},
			_Gov_RuleConditional{
				Field1:    "ID",
				Field2:    "Name",
				Value1:    u.ID,
				Value2:    u.Name,
				Cond:      "John",
				Validator: _Gov_required_if,
			},
			_Gov_RuleRange[int64]{
				Field:     "ID",
				Value:     int64(u.ID),
				Min:       1,
				Max:       1000,
				Validator: _Gov_between_int64,
			},
			_Gov_RuleConditional{
				Field1:    "ID",
				Field2:    "ID2",
				Value1:    u.ID,
				Value2:    u.ID2,
				Validator: _Gov_different,
			},
			_Gov_RuleValueConstraint[int64]{
				Field:     "ID",
				Value:     int64(u.ID),
				Cond:      10,
				Validator: _Gov_size_int64,
			},
			_Gov_RuleConditional{
				Field1:    "ID",
				Field2:    "ID3",
				Value1:    u.ID,
				Value2:    u.ID3,
				Validator: _Gov_same,
			},
			_Gov_RulePresence[string]{
				Field:     "Name",
				Value:     string(u.Name),
				Validator: _Gov_required_string,
			},
			_Gov_RuleValueConstraint[string]{
				Field:     "Email",
				Value:     string(u.Email),
				Validator: _Gov_email_string,
			},
		},
	}
}

func (s UserSchema) Validate() (messages []string) {
	for _, rule := range s.rules {
		if err := rule.Validate(); err != nil {
			messages = append(messages, err.Error())
		}
	}
	return messages
}