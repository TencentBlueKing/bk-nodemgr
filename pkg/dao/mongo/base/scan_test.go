/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package base

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daomongo "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOrm_Scan(t *testing.T) {
	tests := []struct {
		name        string
		limit       int64
		batches     []int
		queryLimits []int64
	}{
		{name: "below batch size", limit: 499, batches: []int{499}, queryLimits: []int64{499}},
		{name: "exact batch size", limit: 500, batches: []int{500}, queryLimits: []int64{500}},
		{name: "partial final batch", limit: 501, batches: []int{500, 1}, queryLimits: []int64{500, 1}},
		{name: "exact multiple", limit: 1000, batches: []int{500, 500}, queryLimits: []int64{500, 500}},
		{name: "fewer matches", limit: 1000, batches: []int{500, 1}, queryLimits: []int64{500, 500}},
		{name: "no matches", limit: 1000, batches: []int{0}, queryLimits: []int64{500}},
		{name: "existing unlimited scan", limit: 0, batches: []int{500, 1}, queryLimits: []int64{500, 500}},
		{name: "unlimited exact multiple", limit: 0, batches: []int{500, 500, 0}, queryLimits: []int64{500, 500, 500}},
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, tt := range tests {
		mt.Run(tt.name, func(mt *mtest.T) {
			total := 0
			for _, count := range tt.batches {
				documents := make([]bson.D, count)
				for i := range count {
					id := total + i + 1
					documents[i] = bson.D{
						{Key: "_id", Value: int64(id)},
						{Key: "data", Value: bson.D{{Key: "id", Value: strconv.Itoa(id)}}},
					}
				}
				total += count
				mt.AddMockResponses(mtest.CreateCursorResponse(0, mt.DB.Name()+"."+mt.Coll.Name(), mtest.FirstBatch, documents...))
			}
			tableName := mt.Coll.Name() + "_" + uuid.NewString()
			orm := NewOrm[*TestData, TestData](&TestDao{collection: mt.Coll, tableName: tableName})
			nCtx := contextx.New(context.Background())
			filter := AliveFilter()
			var data []*TestData
			var err error
			if tt.limit == 0 {
				data, err = orm.ScanAll(nCtx, filter)
			} else {
				data, err = orm.ScanWithLimit(nCtx, filter, tt.limit)
			}
			require.NoError(mt, err)
			require.Len(mt, data, total)
			for i, item := range data {
				require.Equal(mt, strconv.Itoa(i+1), item.ID)
			}
			require.Equal(mt, AliveFilter(), filter)
			for i, limit := range tt.queryLimits {
				event := mt.GetStartedEvent()
				require.NotNil(mt, event)
				require.Equal(mt, "find", event.CommandName)
				require.Equal(mt, limit, event.Command.Lookup("limit").Int64())
				require.Equal(mt, int32(1), event.Command.Lookup("sort").Document().Lookup("_id").Int32())
				if i == 0 {
					continue
				}
				conditions, err := event.Command.Lookup("filter").Document().Lookup("$and").Array().Values()
				require.NoError(mt, err)
				cursor := conditions[len(conditions)-1].Document().Lookup("_id").Document().Lookup("$gt").Int64()
				require.Equal(mt, int64(i*500), cursor)
			}
			require.Nil(mt, mt.GetStartedEvent(), "scan must not issue additional queries after completion")
			operation := "scan_with_limit"
			if tt.limit == 0 {
				operation = "scan_all"
			}
			assertScanMetrics(mt.T, tableName, operation, true, total)
		})
	}
}

func TestOrm_ScanWithLimitRejectsNonPositiveLimit(t *testing.T) {
	for _, limit := range []int64{0, -1} {
		tableName := t.Name() + "_" + uuid.NewString()
		orm := NewOrm[*TestData, TestData](&TestDao{tableName: tableName})
		data, err := orm.ScanWithLimit(contextx.New(context.Background()), nil, limit)
		require.Error(t, err)
		require.Nil(t, data)
		assertScanMetrics(t, tableName, "scan_with_limit", false, 0)
	}
}

func assertScanMetrics(t *testing.T, tableName, operation string, success bool, count int) {
	t.Helper()
	metrics, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	values := make(map[string]float64)
	for _, family := range metrics {
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["table"] != tableName {
				continue
			}
			require.Equal(t, operation, labels["operation"])
			require.Equal(t, strconv.FormatBool(success), labels["success"])
			values[family.GetName()] = metric.GetCounter().GetValue()
			if family.GetName() == "mongodb_request_duration" {
				require.EqualValues(t, 1, metric.GetHistogram().GetSampleCount())
			}
		}
	}
	require.Contains(t, values, "mongodb_request_duration")
	require.Equal(t, float64(1), values["mongodb_request_total"])
	require.Equal(t, float64(count), values["mongodb_response_data_length"])
}

func TestOrm_ScanOperationFailureMetrics(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, operation := range []string{"scan_all", "scan_with_limit"} {
		mt.Run(operation, func(mt *mtest.T) {
			mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 2, Message: "invalid query"}))
			tableName := mt.Coll.Name() + "_" + uuid.NewString()
			orm := NewOrm[*TestData, TestData](&TestDao{collection: mt.Coll, tableName: tableName})
			nCtx := contextx.New(context.Background())
			var err error
			if operation == "scan_all" {
				_, err = orm.ScanAll(nCtx, bson.D{})
			} else {
				_, err = orm.ScanWithLimit(nCtx, bson.D{}, 1)
			}
			require.ErrorContains(mt, err, "invalid query")
			assertScanMetrics(mt.T, tableName, operation, false, 0)
		})
	}
}

func TestOrm_ScanSlowQueryOperation(t *testing.T) {
	// Delay the mock command to exercise the public operation's slow-query event.
	commandMonitor := &event.CommandMonitor{Started: func(context.Context, *event.CommandStartedEvent) {
		time.Sleep(daomongo.DefaultSlowTime + 10*time.Millisecond)
	}}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).
		ClientOptions(options.Client().SetMonitor(commandMonitor)))
	for _, operation := range []string{"scan_all", "scan_with_limit"} {
		mt.Run(operation, func(mt *mtest.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
			defer func() { require.NoError(mt, provider.Shutdown(context.Background())) }()
			ctx, span := provider.Tracer("scan-test").Start(context.Background(), "scan")
			mt.AddMockResponses(mtest.CreateCursorResponse(0, mt.DB.Name()+"."+mt.Coll.Name(), mtest.FirstBatch))
			tableName := mt.Coll.Name() + "_" + uuid.NewString()
			orm := NewOrm[*TestData, TestData](&TestDao{collection: mt.Coll, tableName: tableName})
			var err error
			if operation == "scan_all" {
				_, err = orm.ScanAll(contextx.New(ctx), bson.D{})
			} else {
				_, err = orm.ScanWithLimit(contextx.New(ctx), bson.D{}, 1)
			}
			require.NoError(mt, err)
			span.End()
			spans := recorder.Ended()
			require.Len(mt, spans, 1)
			events := spans[0].Events()
			require.Len(mt, events, 1)
			require.Equal(mt, spanEventSlowQuery, events[0].Name)
			require.Contains(mt, events[0].Attributes, attribute.String(attrKeyORMOperation, operation))
			assertScanMetrics(mt.T, tableName, operation, true, 0)
		})
	}
}

// TestOrm_ScanAll tests the ScanAll method
func TestOrm_ScanAll(t *testing.T) {
	orm, _ := testClient(t)
	nCtx := contextx.New(context.Background())

	_, err := orm.ScanAll(nil, bson.D{})
	require.Error(t, err, "ScanAll() expected error with nil context")

	const total = 505
	datas := make([]*TestData, 0, total+1)
	for i := range total {
		datas = append(datas, &TestData{
			ID:      uuid.NewString(),
			Name:    "scan-all-enabled",
			Value:   i,
			Enabled: true,
		})
	}
	datas = append(datas, &TestData{
		ID:      uuid.NewString(),
		Name:    "scan-all-disabled",
		Value:   total,
		Enabled: false,
	})

	if err := orm.CreateMany(nCtx, datas); err != nil {
		t.Fatalf("Failed to create scan all test data: %v", err)
	}

	filter := bson.D{{
		Key: "$and",
		Value: bson.A{
			bson.D{{Key: FieldKeyTestDataEnabled, Value: true}},
		},
	}}
	wantFilter := bson.D{{
		Key: "$and",
		Value: bson.A{
			bson.D{{Key: FieldKeyTestDataEnabled, Value: true}},
		},
	}}

	got, err := orm.ScanAll(nCtx, filter)
	if err != nil {
		t.Fatalf("ScanAll() error = %v", err)
	}
	if len(got) != total {
		t.Fatalf("ScanAll() got count = %v, want %v", len(got), total)
	}
	if !reflect.DeepEqual(filter, wantFilter) {
		t.Fatalf("ScanAll() mutated filter = %#v, want %#v", filter, wantFilter)
	}

	seen := make(map[string]struct{}, len(got))
	for _, data := range got {
		if !data.Enabled {
			t.Fatalf("ScanAll() returned disabled data: %#v", data)
		}
		if _, ok := seen[data.ID]; ok {
			t.Fatalf("ScanAll() returned duplicate ID: %s", data.ID)
		}
		seen[data.ID] = struct{}{}
	}

	projected, err := orm.ScanAll(nCtx, filter, FieldKeyTestDataName)
	if err != nil {
		t.Fatalf("ScanAll() with projection error = %v", err)
	}
	if len(projected) != total {
		t.Fatalf("ScanAll() with projection got count = %v, want %v", len(projected), total)
	}
	for _, data := range projected {
		if data.Name == "" {
			t.Fatal("ScanAll() with projection returned empty name")
		}
	}
}

// TestBuildScanFilter tests the scan filter builder.
func TestBuildScanFilter(t *testing.T) {
	filter := bson.D{{
		Key: "$and",
		Value: bson.A{
			bson.D{{Key: FieldKeyTestDataEnabled, Value: true}},
		},
	}}
	wantFilter := bson.D{{
		Key: "$and",
		Value: bson.A{
			bson.D{{Key: FieldKeyTestDataEnabled, Value: true}},
		},
	}}

	firstFilter := buildScanFilter(filter, nil)
	if !reflect.DeepEqual(firstFilter, filter) {
		t.Fatalf("buildScanFilter() first filter = %#v, want %#v", firstFilter, filter)
	}

	nextFilter := buildScanFilter(filter, "last-id")
	if !reflect.DeepEqual(filter, wantFilter) {
		t.Fatalf("buildScanFilter() mutated filter = %#v, want %#v", filter, wantFilter)
	}
	if len(nextFilter) != 1 || nextFilter[0].Key != "$and" {
		t.Fatalf("buildScanFilter() next filter = %#v, want $and filter", nextFilter)
	}
	conditions, ok := nextFilter[0].Value.(bson.A)
	if !ok {
		t.Fatalf("buildScanFilter() $and value type = %T, want bson.A", nextFilter[0].Value)
	}
	if len(conditions) != 2 {
		t.Fatalf("buildScanFilter() $and count = %v, want 2", len(conditions))
	}
}
