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
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/combine"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOrm_UpdateFieldBatchingOptIn(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, enabled := range []bool{false, true} {
		mt.Run(strconv.FormatBool(enabled), func(mt *mtest.T) {
			dao := &TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}
			orm := NewOrm[*TestData](dao)
			if enabled {
				orm = NewOrm[*TestData](dao, WithUpdateFieldBatching())
			}
			_, batchingEnabled := orm.updateExecutor.(*updateFieldBatcher)
			require.Equal(mt, enabled, batchingEnabled)
			for _, value := range []int{1, 2} {
				mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 2}))
				require.NoError(mt, orm.UpdateField(contextx.Background(), AliveFilter(), FieldKeyTestDataValue, value))
				evt := mt.GetStartedEvent()
				require.NotNil(mt, evt)
				updates, err := evt.Command.Lookup("updates").Array().Values()
				require.NoError(mt, err)
				require.Len(mt, updates, 1)
				update := updates[0].Document()
				require.True(mt, update.Lookup("multi").Boolean())
				require.Equal(mt, int32(value), update.Lookup("u").Document().Lookup("$set").Document().
					Lookup(FieldKeyTestDataValue).Int32())
			}
		})
	}
}

// Hold the first send open while independent callers arrive, then verify that
// later callers are sent in a second batch rather than sharing the first result.
func TestOrm_UpdateFieldBatchingConcurrentCalls(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	releaseSecond := make(chan struct{})
	completed := make(chan struct{}, 2)
	var sends atomic.Int32
	var releaseOnce sync.Once
	t.Cleanup(func() {
		releaseOnce.Do(func() {
			close(release)
			close(releaseSecond)
		})
	})
	monitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			if evt.CommandName != "update" {
				return
			}
			index := sends.Add(1)
			started <- struct{}{}
			if index == 1 {
				<-release
				return
			}
			<-releaseSecond
		},
		Succeeded: func(_ context.Context, evt *event.CommandSucceededEvent) {
			if evt.CommandName == "update" {
				completed <- struct{}{}
			}
		},
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).
		ClientOptions(mongoOptions.Client().SetMonitor(monitor)))
	mt.Run("independent batches may overlap without early completion", func(mt *mtest.T) {
		recorder := tracetest.NewSpanRecorder()
		provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
		defer func() { require.NoError(mt, provider.Shutdown(context.Background())) }()
		tracer := provider.Tracer("batch-test")
		orm := NewOrm[*TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}, WithUpdateFieldBatching())
		const callers = 2
		batcher := orm.updateExecutor.(*updateFieldBatcher)
		// Use the size trigger so this test does not depend on scheduler timing.
		batcher.combined = combine.New(callers, time.Hour, batcher.send)
		mt.AddMockResponses(mtest.CreateSuccessResponse(), mtest.CreateSuccessResponse())
		first := make(chan error, 1)
		go func() {
			ctx, span := tracer.Start(context.Background(), "first")
			defer span.End()
			first <- orm.UpdateField(contextx.FromContext(ctx), AliveFilter(), FieldKeyTestDataValue, 0)
		}()
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			mt.Fatal("first send did not start")
		}

		results := make(chan error, callers)
		for value := range callers {
			go func() {
				ctx, span := tracer.Start(context.Background(), "caller")
				defer span.End()
				results <- orm.UpdateField(contextx.FromContext(ctx), AliveFilter(), FieldKeyTestDataValue, value+1)
			}()
		}
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			mt.Fatal("independent second batch waited for the first batch")
		}
		select {
		case err := <-results:
			mt.Fatalf("a write returned before its batch completed: %v", err)
		default:
		}
		// The mock deployment has one unsynchronized response queue. Both sends
		// have started, but only one may consume a mock response at a time.
		releaseOnce.Do(func() {
			close(release)
			<-completed
			close(releaseSecond)
		})
		require.NoError(mt, <-first)
		for range callers {
			require.NoError(mt, <-results)
		}
		require.NotNil(mt, mt.GetStartedEvent())
		batch := mt.GetStartedEvent()
		require.NotNil(mt, batch)
		require.False(mt, batch.Command.Lookup("ordered").Boolean())
		updates, err := batch.Command.Lookup("updates").Array().Values()
		require.NoError(mt, err)
		require.Len(mt, updates, callers, "same-filter writes must not be merged or dropped")
		values := make(map[int32]bool)
		for _, update := range updates {
			values[update.Document().Lookup("u").Document().Lookup("$set").Document().
				Lookup(FieldKeyTestDataValue).Int32()] = true
		}
		require.Len(mt, values, callers)
		require.Nil(mt, mt.GetStartedEvent())
		calls := make(map[string]sdktrace.ReadOnlySpan)
		var flushes []sdktrace.ReadOnlySpan
		for _, span := range recorder.Ended() {
			switch span.Name() {
			case "mongo.batch.update_many":
				calls[span.SpanContext().SpanID().String()] = span
			case "mongo.batch.flush":
				flushes = append(flushes, span)
			}
		}
		require.Len(mt, calls, callers+1)
		require.Len(mt, flushes, 2)
		for _, flush := range flushes {
			require.Contains(mt, flush.Attributes(), attribute.Int(attrKeyBatchSize, len(flush.Links())))
			for _, link := range flush.Links() {
				call, ok := calls[link.SpanContext.SpanID().String()]
				require.True(mt, ok, "every participating caller must be linked to its send")
				require.Equal(mt, call.SpanContext(), link.SpanContext)
				require.Contains(mt, call.Attributes(), attribute.Int(attrKeyBatchSize, len(flush.Links())))
				delete(calls, link.SpanContext.SpanID().String())
			}
		}
		require.Empty(mt, calls, "no caller may be omitted from its send's links")
	})
}

func TestOrm_UpdateFieldBatchingSpans(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, scenario := range []string{"immediate", "batch", "error", "session"} {
		mt.Run(scenario, func(mt *mtest.T) {
			recorder := tracetest.NewSpanRecorder()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
			defer func() { require.NoError(mt, provider.Shutdown(context.Background())) }()
			ctx, parent := provider.Tracer("batch-test").Start(context.Background(), "caller")
			defer parent.End()
			nCtx := contextx.FromContext(ctx)
			dao := &TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}
			orm := NewOrm[*TestData](dao, WithUpdateFieldBatching())
			if scenario == "immediate" {
				orm = NewOrm[*TestData](dao)
			}
			if scenario == "session" {
				session, err := mt.Client.StartSession()
				require.NoError(mt, err)
				defer session.EndSession(context.Background())
				nCtx = contextx.FromContext(mongo.NewSessionContext(nCtx, session))
			}
			response := mtest.CreateSuccessResponse()
			if scenario == "error" {
				response = mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 2, Message: "invalid query"})
			}
			mt.AddMockResponses(response)
			err := orm.UpdateField(nCtx, AliveFilter(), FieldKeyTestDataValue, 1)
			if scenario == "error" {
				require.ErrorContains(mt, err, "invalid query")
			} else {
				require.NoError(mt, err)
			}
			spans := recorder.Ended()
			if scenario == "immediate" || scenario == "session" {
				require.Empty(mt, spans, "direct writes must not be marked as batched")
				return
			}
			require.Len(mt, spans, 2)
			require.Equal(mt, "mongo.batch.flush", spans[0].Name())
			require.Equal(mt, "mongo.batch.update_many", spans[1].Name())
			require.Equal(mt, parent.SpanContext().SpanID(), spans[1].Parent().SpanID())
			require.Equal(mt, spans[1].SpanContext().SpanID(), spans[0].Parent().SpanID())
			for _, span := range spans {
				require.Contains(mt, span.Attributes(), attribute.String(attrKeyMongoCollection, dao.tableName))
				require.Contains(mt, span.Attributes(), attribute.Int(attrKeyBatchSize, 1))
				if scenario == "error" {
					require.Equal(mt, codes.Error, span.Status().Code)
					require.NotEmpty(mt, span.Events())
				}
			}
		})
	}
}

func TestUpdateFieldBatchErrors(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("indexed write error", func(mt *mtest.T) {
		batcher := newUpdateFieldBatcher(
			&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()},
			&immediateUpdateExecutor{dao: &TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}},
		)
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Index: 1, Code: 121, Message: "invalid field"}))
		items := make([]*pendingFieldUpdate, 3)
		for index := range items {
			items[index] = &pendingFieldUpdate{
				ctx: contextx.Background(),
				model: mongo.NewUpdateManyModel().SetFilter(AliveFilter()).
					SetUpdate(buildUpdateField(FieldKeyTestDataValue, index)),
			}
		}
		results, err := batcher.send("", items)
		require.NoError(mt, err)
		require.NoError(mt, results[0])
		require.NoError(mt, results[2])
		var writeErr mongo.WriteException
		require.ErrorAs(mt, results[1], &writeErr)
		require.Len(mt, writeErr.WriteErrors, 1)
		require.Equal(mt, 121, writeErr.WriteErrors[0].Code)
		require.Equal(mt, 0, writeErr.WriteErrors[0].Index)
	})
	mt.Run("invalid encoding does not enter queue", func(mt *mtest.T) {
		orm := NewOrm[*TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}, WithUpdateFieldBatching())
		require.Error(mt, orm.UpdateField(contextx.Background(), AliveFilter(), FieldKeyTestDataValue, make(chan int)))
		require.Nil(mt, mt.GetStartedEvent())
	})
	mt.Run("command and write concern errors affect every caller", func(mt *mtest.T) {
		for _, err := range []error{
			mongo.CommandError{Code: 13, Message: "not authorized"},
			mongo.BulkWriteException{WriteConcernError: &mongo.WriteConcernError{Code: 64, Message: "timeout"}},
		} {
			for index := range 3 {
				require.Error(mt, fieldUpdateBatchError(err, index))
			}
		}
	})
}

func TestOrm_UpdateFieldBatchingQueuedCancellation(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("skip canceled write and preserve result indexes", func(mt *mtest.T) {
		orm := NewOrm[*TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}, WithUpdateFieldBatching())
		batcher := orm.updateExecutor.(*updateFieldBatcher)
		batcher.combined = combine.New(2, time.Hour, batcher.send)
		mt.AddMockResponses(mtest.CreateSuccessResponse(),
			mtest.CreateWriteErrorsResponse(mtest.WriteError{Index: 0, Code: 121, Message: "live write failed"}))
		// The first call starts immediately. Subsequent calls wait for the size
		// threshold, letting cancellation happen before the next send.
		require.NoError(mt, orm.UpdateField(contextx.Background(), AliveFilter(), FieldKeyTestDataValue, 0))
		nCtx, cancel := contextx.WithTimeout(contextx.Background(), 10*time.Millisecond)
		defer cancel()
		require.ErrorIs(mt, orm.UpdateField(nCtx, AliveFilter(), FieldKeyTestDataValue, 1), context.DeadlineExceeded)
		require.ErrorContains(mt, orm.UpdateField(contextx.Background(), AliveFilter(), FieldKeyTestDataValue, 2),
			"live write failed")
		require.NotNil(mt, mt.GetStartedEvent())
		batch := mt.GetStartedEvent()
		require.NotNil(mt, batch)
		updates, err := batch.Command.Lookup("updates").Array().Values()
		require.NoError(mt, err)
		require.Len(mt, updates, 1, "canceled write must not be sent")
		require.Equal(mt, int32(2), updates[0].Document().Lookup("u").Document().Lookup("$set").Document().
			Lookup(FieldKeyTestDataValue).Int32())
		require.Nil(mt, mt.GetStartedEvent())
	})
}

func TestOrm_UpdateFieldBatchingDispatchedCancellation(t *testing.T) {
	nCtx, cancel := contextx.WithCancel(contextx.Background())
	defer cancel()
	monitor := &event.CommandMonitor{Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
		if evt.CommandName != "update" {
			return
		}
		cancel()
		require.NoError(t, ctx.Err(), "one caller must not cancel a shared send")
		_, bounded := ctx.Deadline()
		require.True(t, bounded, "shared sends need an independent timeout")
	}}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).
		ClientOptions(mongoOptions.Client().SetMonitor(monitor)))
	mt.Run("wait for dispatched write", func(mt *mtest.T) {
		orm := NewOrm[*TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}, WithUpdateFieldBatching())
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		require.NoError(mt, orm.UpdateField(nCtx, AliveFilter(), FieldKeyTestDataValue, 1))
		require.ErrorIs(mt, nCtx.Err(), context.Canceled)
	})
}

func TestOrm_UpdateFieldBatchingSessionBypass(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("session keeps direct UpdateMany", func(mt *mtest.T) {
		orm := NewOrm[*TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}, WithUpdateFieldBatching())
		session, err := mt.Client.StartSession()
		require.NoError(mt, err)
		defer session.EndSession(context.Background())
		nCtx := contextx.FromContext(mongo.NewSessionContext(contextx.Background(), session))
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		require.NoError(mt, orm.UpdateField(nCtx, AliveFilter(), FieldKeyTestDataValue, 1))
		evt := mt.GetStartedEvent()
		require.NotNil(mt, evt)
		require.True(mt, evt.Command.Lookup("ordered").Boolean())
	})
}
