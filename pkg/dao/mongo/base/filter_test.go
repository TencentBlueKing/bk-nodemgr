package base

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWithGreaterThanValue(t *testing.T) {
	tests := []struct {
		name     string
		filter   bson.D
		opt      OptFn
		expected bson.D
	}{
		{
			name:     "string value",
			filter:   bson.D{},
			opt:      WithGreaterThanValue("data.dynamic.agent_id", ""),
			expected: bson.D{{Key: "data.dynamic.agent_id", Value: bson.M{"$gt": ""}}},
		},
		{
			name:     "int64 value",
			filter:   bson.D{},
			opt:      WithGreaterThanValue("data.host_id", int64(1000)),
			expected: bson.D{{Key: "data.host_id", Value: bson.M{"$gt": int64(1000)}}},
		},
		{
			name:   "append to existing filter",
			filter: bson.D{{Key: "basic.is_deleted", Value: false}},
			opt:    WithGreaterThanValue("data.dynamic.agent_id", ""),
			expected: bson.D{
				{Key: "basic.is_deleted", Value: false},
				{Key: "data.dynamic.agent_id", Value: bson.M{"$gt": ""}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.opt(tt.filter)
			if !reflect.DeepEqual(filter, tt.expected) {
				t.Fatalf("filter = %#v, want %#v", filter, tt.expected)
			}
		})
	}
}

func TestWithFuzzyValues(t *testing.T) {
	tests := []struct {
		name     string
		filter   bson.D
		opt      OptFn
		expected bson.D
	}{
		{
			name:   "escape unmatched ascii parenthesis with fullwidth parenthesis",
			filter: bson.D{},
			opt:    WithFuzzyValues("data.static.dept_name", "安全产品三部(New）"),
			expected: bson.D{
				{Key: "data.static.dept_name", Value: bson.M{"$regex": "(安全产品三部\\(New）)", "$options": "i"}},
			},
		},
		{
			name:   "escape values while preserving or between values",
			filter: bson.D{},
			opt:    WithFuzzyValues("data.name", "a.*", "b|c"),
			expected: bson.D{
				{Key: "data.name", Value: bson.M{"$regex": "(a\\.\\*|b\\|c)", "$options": "i"}},
			},
		},
		{
			name:     "empty values no-op",
			filter:   bson.D{{Key: "basic.is_deleted", Value: false}},
			opt:      WithFuzzyValues("data.name"),
			expected: bson.D{{Key: "basic.is_deleted", Value: false}},
		},
		{
			name:   "append to existing filter",
			filter: bson.D{{Key: "basic.is_deleted", Value: false}},
			opt:    WithFuzzyValues("data.name", "node(1)"),
			expected: bson.D{
				{Key: "basic.is_deleted", Value: false},
				{Key: "data.name", Value: bson.M{"$regex": "(node\\(1\\))", "$options": "i"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.opt(tt.filter)
			if !reflect.DeepEqual(filter, tt.expected) {
				t.Fatalf("filter = %#v, want %#v", filter, tt.expected)
			}
		})
	}
}

func TestWithoutFuzzyValues(t *testing.T) {
	tests := []struct {
		name     string
		filter   bson.D
		opt      OptFn
		expected bson.D
	}{
		{
			name:   "escape unmatched ascii parenthesis with fullwidth parenthesis",
			filter: bson.D{},
			opt:    WithoutFuzzyValues("data.static.dept_name", "安全产品三部(New）"),
			expected: bson.D{
				{Key: "data.static.dept_name", Value: bson.M{"$not": bson.M{"$regex": "(安全产品三部\\(New）)", "$options": "i"}}},
			},
		},
		{
			name:   "escape values while preserving or between values",
			filter: bson.D{},
			opt:    WithoutFuzzyValues("data.name", "a.*", "b|c"),
			expected: bson.D{
				{Key: "data.name", Value: bson.M{"$not": bson.M{"$regex": "(a\\.\\*|b\\|c)", "$options": "i"}}},
			},
		},
		{
			name:     "empty values no-op",
			filter:   bson.D{{Key: "basic.is_deleted", Value: false}},
			opt:      WithoutFuzzyValues("data.name"),
			expected: bson.D{{Key: "basic.is_deleted", Value: false}},
		},
		{
			name:   "append to existing filter",
			filter: bson.D{{Key: "basic.is_deleted", Value: false}},
			opt:    WithoutFuzzyValues("data.name", "node(1)"),
			expected: bson.D{
				{Key: "basic.is_deleted", Value: false},
				{Key: "data.name", Value: bson.M{"$not": bson.M{"$regex": "(node\\(1\\))", "$options": "i"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.opt(tt.filter)
			if !reflect.DeepEqual(filter, tt.expected) {
				t.Fatalf("filter = %#v, want %#v", filter, tt.expected)
			}
		})
	}
}
