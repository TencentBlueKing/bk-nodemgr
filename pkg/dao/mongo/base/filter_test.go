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
