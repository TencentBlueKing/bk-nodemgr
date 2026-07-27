package host

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestWithDynamicAgentIDNotEmpty(t *testing.T) {
	filter := WithDynamicAgentIDNotEmpty()(bson.D{})
	expected := bson.D{{Key: FieldKeyDynamicAgentID, Value: bson.M{"$gt": ""}}}

	if !reflect.DeepEqual(filter, expected) {
		t.Fatalf("filter = %#v, want %#v", filter, expected)
	}
}
