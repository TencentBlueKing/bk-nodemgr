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
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daomongo "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/singleflight"
)

const (
	// Bound each send without adding a timer or a background worker.
	updateFieldBatchLimit = 128
	// A shared send must not inherit one caller's cancellation or run indefinitely.
	updateFieldBatchTimeout = 30 * time.Second
)

type ormOptions struct {
	updateFieldBatching bool
}

// OrmOption configures optional ORM behavior at construction time.
type OrmOption func(*ormOptions)

// WithUpdateFieldBatching batches overlapping UpdateField calls on one ORM instance.
// Each call remains an independent UpdateMany and waits for its own write result.
// Concurrent writes are unordered; sequential calls remain sequential. Sessions
// bypass batching. Cancellation before dispatch skips the write; after dispatch,
// callers wait for the shared send (bounded by 30 seconds), which may still commit.
// Sends use short-lived singleflight goroutines, not a background service. No
// retries, field merging, or cross-instance batching is used.
func WithUpdateFieldBatching() OrmOption {
	return func(options *ormOptions) {
		options.updateFieldBatching = true
	}
}

type pendingFieldUpdate struct {
	ctx        contextx.IContext
	model      mongo.WriteModel
	done       chan struct{}
	err        error // Published by closing done; read only after receiving from done.
	dispatched bool  // Protected by updateFieldBatcher.mu.
}

type updateFieldBatcher struct {
	dao     IDao
	mu      sync.Mutex
	pending []*pendingFieldUpdate
	sends   singleflight.Group
}

func (batcher *updateFieldBatcher) update(nCtx contextx.IContext, filter bson.D, field string, value any) error {
	if err := nCtx.Err(); err != nil {
		return err
	}
	// Encode before admission so a caller's invalid value cannot fail its peers.
	encodedFilter, err := bson.Marshal(filter)
	if err != nil {
		return fmt.Errorf("encode batched update filter: %w", err)
	}
	encodedUpdate, err := bson.Marshal(buildUpdateField(field, value))
	if err != nil {
		return fmt.Errorf("encode batched update fields: %w", err)
	}
	item := &pendingFieldUpdate{
		ctx: nCtx,
		model: mongo.NewUpdateManyModel().SetFilter(bson.Raw(encodedFilter)).
			SetUpdate(bson.Raw(encodedUpdate)).SetUpsert(false),
		done: make(chan struct{}),
	}
	batcher.mu.Lock()
	batcher.pending = append(batcher.pending, item)
	batcher.mu.Unlock()

	for {
		select {
		case <-item.done:
			return item.err
		default:
		}
		// Joining a flight does not mean this item was in its drained batch.
		// Recheck item.done and participate in another send when necessary.
		resultCh := batcher.sends.DoChan("update_field", func() (any, error) {
			batcher.flush()

			return nil, nil // nolint: nilnil // Results belong to individual items, not the flight.
		})
		select {
		case <-item.done:
			return item.err
		case <-nCtx.Done():
			batcher.mu.Lock()
			if !item.dispatched {
				index := slices.Index(batcher.pending, item)
				batcher.pending = slices.Delete(batcher.pending, index, index+1)
				batcher.mu.Unlock()

				return nCtx.Err()
			}
			batcher.mu.Unlock()
			<-item.done

			return item.err
		case result := <-resultCh:
			if result.Err != nil {
				return result.Err
			}
		}
	}
}

func (batcher *updateFieldBatcher) flush() {
	batcher.mu.Lock()
	count := min(len(batcher.pending), updateFieldBatchLimit)
	batch := batcher.pending[:count:count]
	batcher.pending = batcher.pending[count:]
	for _, item := range batch {
		item.dispatched = true
	}
	batcher.mu.Unlock()

	items := make([]*pendingFieldUpdate, 0, len(batch))
	models := make([]mongo.WriteModel, 0, len(batch))
	for _, item := range batch {
		if err := item.ctx.Err(); err != nil {
			item.err = err
			close(item.done)

			continue
		}
		items = append(items, item)
		models = append(models, item.model)
	}
	if len(items) == 0 {
		return
	}

	nCtx, cancel := contextx.WithTimeout(contextx.WithoutCancel(items[0].ctx), updateFieldBatchTimeout)
	defer cancel()
	metric := newMetricData(batcher.dao.GetTableName()).start(daomongo.MetricOperationBulkWrite, len(models))
	result, err := batcher.dao.GetClient().BulkWrite(nCtx, models, mongoOptions.BulkWrite().SetOrdered(false))
	modified := 0
	if result != nil {
		modified = int(result.ModifiedCount)
	}
	metric.end(err, modified)
	for index, item := range items {
		item.err = fieldUpdateBatchError(err, index)
		close(item.done)
	}
}

func fieldUpdateBatchError(err error, index int) error {
	var bulkErr mongo.BulkWriteException
	if !errors.As(err, &bulkErr) {
		return err
	}
	// Command/transport errors affect the whole batch. Indexed write errors only
	// affect their own caller; preserve UpdateMany's WriteException error type.
	writeErr := mongo.WriteException{WriteConcernError: bulkErr.WriteConcernError, Labels: bulkErr.Labels}
	for _, failure := range bulkErr.WriteErrors {
		if failure.Index != index {
			continue
		}
		failure.Index = 0
		writeErr.WriteErrors = append(writeErr.WriteErrors, failure.WriteError)
	}
	if writeErr.WriteConcernError != nil || len(writeErr.WriteErrors) > 0 {
		return writeErr
	}

	return nil
}
