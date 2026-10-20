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
	"testing"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
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
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	monitor := &event.CommandMonitor{Started: func(_ context.Context, evt *event.CommandStartedEvent) {
		if evt.CommandName == "update" {
			once.Do(func() {
				close(started)
				<-release
			})
		}
	}}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).
		ClientOptions(mongoOptions.Client().SetMonitor(monitor)))
	mt.Run("late arrivals and canceled callers", func(mt *mtest.T) {
		orm := NewOrm[*TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}, WithUpdateFieldBatching())
		mt.AddMockResponses(mtest.CreateSuccessResponse(), mtest.CreateSuccessResponse())
		first := make(chan error, 1)
		go func() {
			first <- orm.UpdateField(contextx.Background(), AliveFilter(), FieldKeyTestDataValue, 0)
		}()
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			mt.Fatal("first send did not start")
		}

		const callers = 12
		results := make(chan error, callers)
		for value := range callers {
			go func() {
				results <- orm.UpdateField(contextx.Background(), AliveFilter(), FieldKeyTestDataValue, value+1)
			}()
		}
		canceledCtx, cancel := contextx.WithCancel(contextx.Background())
		defer cancel()
		canceled := make(chan error, 1)
		go func() {
			canceled <- orm.UpdateField(canceledCtx, AliveFilter(), FieldKeyTestDataName, "canceled")
		}()
		require.Eventually(mt, func() bool {
			batcher := orm.updateExecutor.(*updateFieldBatcher)
			batcher.mu.Lock()
			defer batcher.mu.Unlock()

			return len(batcher.pending) == callers+1
		}, 5*time.Second, time.Millisecond)
		cancel()
		select {
		case err := <-canceled:
			require.ErrorIs(mt, err, context.Canceled)
		case <-time.After(5 * time.Second):
			mt.Fatal("queued cancellation waited for an unrelated send")
		}
		select {
		case err := <-results:
			mt.Fatalf("a queued write returned before being sent: %v", err)
		default:
		}
		releaseOnce.Do(func() { close(release) })
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
	})
}

func TestUpdateFieldBatchErrors(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("indexed write error", func(mt *mtest.T) {
		batcher := &updateFieldBatcher{
			dao:       &TestDao{collection: mt.Coll, tableName: mt.Coll.Name()},
			immediate: &immediateUpdateExecutor{dao: &TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}},
		}
		for range 3 {
			batcher.pending = append(batcher.pending, &pendingFieldUpdate{
				ctx: contextx.Background(), done: make(chan struct{}),
				model: mongo.NewUpdateManyModel().SetFilter(AliveFilter()).
					SetUpdate(buildUpdateField(FieldKeyTestDataValue, 1)),
			})
		}
		items := append([]*pendingFieldUpdate(nil), batcher.pending...)
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Index: 1, Code: 121, Message: "invalid field"}))
		batcher.flush()
		require.NoError(mt, items[0].err)
		require.NoError(mt, items[2].err)
		var writeErr mongo.WriteException
		require.ErrorAs(mt, items[1].err, &writeErr)
		require.Len(mt, writeErr.WriteErrors, 1)
		require.Equal(mt, 121, writeErr.WriteErrors[0].Code)
		require.Equal(mt, 0, writeErr.WriteErrors[0].Index)
	})
	mt.Run("invalid encoding does not enter queue", func(mt *mtest.T) {
		orm := NewOrm[*TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()}, WithUpdateFieldBatching())
		require.Error(mt, orm.UpdateField(contextx.Background(), AliveFilter(), FieldKeyTestDataValue, make(chan int)))
		batcher := orm.updateExecutor.(*updateFieldBatcher)
		require.Empty(mt, batcher.pending)
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
		batcher := orm.updateExecutor.(*updateFieldBatcher)
		require.Empty(mt, batcher.pending)
	})
}
