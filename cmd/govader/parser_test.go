package main

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test__parseSchema(t *testing.T) {
	t.Parallel()
	tests := [...]struct {
		name    string
		info    []StructInfo
		want    []Schema
		wantErr bool
	}{
		{
			name: "parse presence rule",
			info: []StructInfo{
				{
					Name: "User",
					FieldList: []FieldInfo{
						{Name: "ID", Tag: "required", Type: types.IsInteger},
						{Name: "Name", Tag: "required", Type: types.IsString},
					},
				},
			},
			want: []Schema{
				{
					Rules: []SchemaRule{
						{Name: "required", Type: rulePresence, Field1: "ID", Cond1: &Value{Type: types.IsInteger}},
						{Name: "required", Type: rulePresence, Field1: "Name", Cond1: &Value{Type: types.IsString}},
					},
					Validators: []string{"required"},
				},
			},
		},
		{
			name: "parse value constraint rule",
			info: []StructInfo{
				{
					Name: "User",
					FieldList: []FieldInfo{
						{Name: "ID", Tag: "min=1", Type: types.IsInteger},
						{Name: "Name", Tag: "size=10", Type: types.IsInteger},
						{Name: "Age", Tag: "regexp=^[0-9]*$", Type: types.IsString},
						{Name: "Email", Tag: "email", Type: types.IsString},
					},
				},
			},
			want: []Schema{
				{
					Rules: []SchemaRule{
						{Name: "min", Type: ruleValueConstraint, Field1: "ID", Cond1: &Value{Value: int64(1), Type: types.IsInteger}},
						{Name: "size", Type: ruleValueConstraint, Field1: "Name", Cond1: &Value{Value: int64(10), Type: types.IsInteger}},
						{Name: "regexp", Type: ruleValueConstraint, Field1: "Age", Cond1: &Value{Value: "^[0-9]*$", Type: types.IsString}},
						{Name: "email", Type: ruleValueConstraint, Field1: "Email", Cond1: &Value{Type: types.IsString}},
					},
					Validators: []string{"min", "size", "regexp", "email"},
				},
			},
		},
		{
			name: "parse range rule",
			info: []StructInfo{
				{
					Name: "User",
					FieldList: []FieldInfo{
						{Name: "Age", Tag: "between=1,10", Type: types.IsInteger},
					},
				},
			},
			want: []Schema{
				{
					Rules: []SchemaRule{
						{
							Name:   "between",
							Type:   ruleRange,
							Field1: "Age",
							Cond1:  &Value{Value: int64(1), Type: types.IsInteger},
							Cond2:  &Value{Value: int64(10), Type: types.IsInteger},
						},
					},
					Validators: []string{"between"},
				},
			},
		},
		{
			name: "parse conditional rule",
			info: []StructInfo{
				{
					Name: "User",
					FieldList: []FieldInfo{
						{Name: "ID", Tag: "required_if:Name=John;different:ID2;same:ID3;required_with:ID1"},
					},
				},
			},
			want: []Schema{
				{
					Rules: []SchemaRule{
						{Name: "required_if", Type: ruleConditional, Field1: "ID", Field2: "Name", Cond1: &Value{Value: "John"}},
						{Name: "different", Type: ruleConditional, Field1: "ID", Field2: "ID2"},
						{Name: "same", Type: ruleConditional, Field1: "ID", Field2: "ID3"},
						{Name: "required_with", Type: ruleConditional, Field1: "ID", Field2: "ID1"},
					},
					Validators: []string{"required_if", "different", "same", "required_with"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemas, err := parseSchema(tt.info)
			if tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.NoError(t, err)
				for i, want := range tt.want {
					got := schemas[i]
					assert.Equal(t, want.Rules, got.Rules)
					for _, v := range want.Validators {
						assert.Contains(t, got.Validators, v)
					}
				}
			}
		})
	}
}
