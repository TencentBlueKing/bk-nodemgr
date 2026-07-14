/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package base

import (
	"context"
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

type aggregateDistinctTestDao struct {
	collection *mongo.Collection
}

func (d *aggregateDistinctTestDao) GetClient() *mongo.Collection {
	return d.collection
}

func (d *aggregateDistinctTestDao) GetTableName() string {
	return d.collection.Name()
}

func (d *aggregateDistinctTestDao) GetIndexes() []mongo.IndexModel {
	return nil
}

func TestAggregateDistinct(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("one aggregate command", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			0,
			mt.DB.Name()+"."+mt.Coll.Name(),
			mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: nil},
				{Key: "values", Value: bson.A{"", "agent", int32(1)}},
			},
		))

		dao := &aggregateDistinctTestDao{collection: mt.Coll}
		filter := AliveFilter()
		fields := []AggregateDistinctField{{
			ResultKey: "values",
			FieldPath: "data.dynamic.node_role",
			BSONType:  bson.TypeString,
		}}
		result, err := AggregateDistinct(contextx.New(context.Background()), dao, filter, fields)
		if err != nil {
			mt.Fatalf("AggregateDistinct() error = %v", err)
		}
		assertRawStrings(mt.T, result["values"], []string{"", "agent"})

		event := mt.GetStartedEvent()
		if event == nil || event.CommandName != "aggregate" {
			mt.Fatalf("started event = %#v, want one aggregate command", event)
		}
		if mt.GetStartedEvent() != nil {
			mt.Fatal("AggregateDistinct() executed more than one command")
		}

		allowDiskUse, err := event.Command.LookupErr("allowDiskUse")
		if err != nil || allowDiskUse.Type != bson.TypeBoolean || !allowDiskUse.Boolean() {
			mt.Fatalf("allowDiskUse = %v, error = %v, want true", allowDiskUse, err)
		}
		if _, err := event.Command.LookupErr("hint"); err == nil {
			mt.Fatal("aggregate command unexpectedly contains hint")
		}

		pipeline := event.Command.Lookup("pipeline").Array()
		stages, err := pipeline.Values()
		if err != nil {
			mt.Fatalf("decode pipeline error = %v", err)
		}
		if len(stages) != 2 {
			mt.Fatalf("pipeline stage count = %d, want 2", len(stages))
		}
		match := stages[0].Document().Lookup("$match").Document()
		if !reflect.DeepEqual(match, marshalRawDocument(mt.T, filter)) {
			mt.Fatalf("$match = %s, want %s", match, marshalRawDocument(mt.T, filter))
		}
		group := stages[1].Document().Lookup("$group").Document()
		if group.Lookup("_id").Type != bson.TypeNull {
			mt.Fatalf("$group._id type = %s, want null", group.Lookup("_id").Type)
		}
		addToSet := group.Lookup("values").Document().Lookup("$addToSet").StringValue()
		if addToSet != "$data.dynamic.node_role" {
			mt.Fatalf("$addToSet = %q", addToSet)
		}
	})
}

func TestAggregateDistinctRejectsMultipleDocuments(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("multiple group documents", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(
			0,
			mt.DB.Name()+"."+mt.Coll.Name(),
			mtest.FirstBatch,
			bson.D{{Key: "values", Value: bson.A{"one"}}},
			bson.D{{Key: "values", Value: bson.A{"two"}}},
		))

		dao := &aggregateDistinctTestDao{collection: mt.Coll}
		_, err := AggregateDistinct(
			contextx.New(context.Background()),
			dao,
			AliveFilter(),
			[]AggregateDistinctField{{ResultKey: "values", BSONType: bson.TypeString}},
		)
		if err == nil {
			mt.Fatal("AggregateDistinct() error = nil, want multiple document error")
		}
	})
}

func TestAggregateDistinctEmptyFields(t *testing.T) {
	result, err := AggregateDistinct(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("AggregateDistinct(empty fields) error = %v", err)
	}
	if result == nil || len(result) != 0 {
		t.Fatalf("AggregateDistinct(empty fields) result = %#v, want non-nil empty map", result)
	}
}

func TestBuildAggregateDistinctPipeline(t *testing.T) {
	filter := bson.D{
		{Key: FieldKeyIsDeleted, Value: false},
		{Key: "data.tenant_id", Value: "tenant"},
	}
	fields := []AggregateDistinctField{
		{ResultKey: "biz_id", FieldPath: "data.static.biz_id", BSONType: bson.TypeInt64},
		{ResultKey: "node_role", FieldPath: "data.dynamic.node_role", BSONType: bson.TypeString},
	}

	got := buildAggregateDistinctPipeline(filter, fields)
	want := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "biz_id", Value: bson.D{{Key: "$addToSet", Value: "$data.static.biz_id"}}},
			{Key: "node_role", Value: bson.D{{Key: "$addToSet", Value: "$data.dynamic.node_role"}}},
		}}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildAggregateDistinctPipeline() = %#v, want %#v", got, want)
	}

	opts := aggregateDistinctOptions()
	if opts.AllowDiskUse == nil || !*opts.AllowDiskUse {
		t.Fatal("aggregate distinct must enable allowDiskUse")
	}
	if opts.Hint != nil {
		t.Fatalf("aggregate distinct hint = %#v, want nil", opts.Hint)
	}
}

func TestDecodeAggregateDistinctResult(t *testing.T) {
	fields := []AggregateDistinctField{
		{ResultKey: "strings", BSONType: bson.TypeString},
		{ResultKey: "ids", BSONType: bson.TypeInt64},
		{ResultKey: "missing", BSONType: bson.TypeString},
		{ResultKey: "null", BSONType: bson.TypeString},
	}
	document := marshalRawDocument(t, bson.D{
		{Key: "strings", Value: bson.A{"", "agent", int64(1), int32(2), 3.0, nil, bson.A{"nested"}}},
		{Key: "ids", Value: bson.A{int64(0), int64(42), int32(7), 8.0, "9", bson.A{int64(10)}}},
		{Key: "null", Value: nil},
	})

	result, err := decodeAggregateDistinctResult(document, fields)
	if err != nil {
		t.Fatalf("decodeAggregateDistinctResult() error = %v", err)
	}

	assertRawStrings(t, result["strings"], []string{"", "agent"})
	assertRawInt64s(t, result["ids"], []int64{0, 42})
	for _, key := range []string{"missing", "null"} {
		if result[key] == nil {
			t.Fatalf("result[%q] is nil, want non-nil empty slice", key)
		}
		if len(result[key]) != 0 {
			t.Fatalf("result[%q] = %#v, want empty slice", key, result[key])
		}
	}
}

func TestDecodeAggregateDistinctResultRejectsNonArray(t *testing.T) {
	fields := []AggregateDistinctField{{ResultKey: "values", BSONType: bson.TypeString}}
	document := marshalRawDocument(t, bson.D{{Key: "values", Value: "not-an-array"}})

	if _, err := decodeAggregateDistinctResult(document, fields); err == nil {
		t.Fatal("decodeAggregateDistinctResult() error = nil, want non-array error")
	}
}

func TestNewAggregateDistinctResult(t *testing.T) {
	result := newAggregateDistinctResult([]AggregateDistinctField{
		{ResultKey: "selected", BSONType: bson.TypeString},
	})

	if result["selected"] == nil {
		t.Fatal("selected result is nil, want non-nil empty slice")
	}
	if _, ok := result["unselected"]; ok {
		t.Fatal("unselected result key exists")
	}
}

func marshalRawDocument(t *testing.T, document bson.D) bson.Raw {
	t.Helper()

	data, err := bson.Marshal(document)
	if err != nil {
		t.Fatalf("bson.Marshal() error = %v", err)
	}

	return data
}

func assertRawStrings(t *testing.T, values []bson.RawValue, expected []string) {
	t.Helper()

	got := make([]string, len(values))
	for idx, value := range values {
		got[idx] = value.StringValue()
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("string values = %#v, want %#v", got, expected)
	}
}

func assertRawInt64s(t *testing.T, values []bson.RawValue, expected []int64) {
	t.Helper()

	got := make([]int64, len(values))
	for idx, value := range values {
		got[idx] = value.Int64()
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("int64 values = %#v, want %#v", got, expected)
	}
}
