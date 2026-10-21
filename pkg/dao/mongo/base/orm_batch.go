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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daomongo "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/combine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Bound each MongoDB send; combine's limit is a launch threshold, not a cap.
	updateFieldBatchLimit = 128
	// Allow overlapping calls to collect without CMDB's longer batching window.
	updateFieldBatchGap = time.Millisecond
	// A shared send must not inherit one caller's cancellation or run indefinitely.
	updateFieldBatchTimeout = 30 * time.Second
)

type ormOptions struct {
	updateFieldBatching bool
}

type updateExecutor interface {
	UpdateMany(nCtx contextx.IContext, filter bson.D, update bson.D) error
}

type immediateUpdateExecutor struct {
	dao IDao
}

func (executor *immediateUpdateExecutor) UpdateMany(
	nCtx contextx.IContext, filter bson.D, update bson.D,
) error {

	metric := newMetricData(executor.dao.GetTableName()).start(daomongo.MetricOperationUpdateMany, len(filter))
	result, err := executor.dao.GetClient().UpdateMany(nCtx, filter, update)
	modified := 0
	if result != nil {
		modified = int(result.ModifiedCount)
	}
	metric.end(err, modified)
	if err != nil {
		return err
	}

	logger.G.Sys().Ctx(nCtx).
		With("table", executor.dao.GetTableName(), "matched-count", result.MatchedCount, "modified-count", result.ModifiedCount).
		Debug("updated documents")

	return nil
}

// OrmOption configures optional ORM behavior at construction time.
type OrmOption func(*ormOptions)

// WithUpdateFieldBatching batches overlapping UpdateField calls on one ORM instance.
// Each call remains an independent UpdateMany and waits for its own write result.
// Concurrent writes are unordered; sequential calls remain sequential. Sessions
// bypass batching. Cancellation before dispatch skips the write; after dispatch,
// callers wait for the shared send (bounded by 30 seconds), which may still commit.
// combine launches batches by size or a 1ms interval; independent batches may
// overlap. No retries, field merging, or cross-instance batching is used.
func WithUpdateFieldBatching() OrmOption {
	return func(options *ormOptions) {
		options.updateFieldBatching = true
	}
}

type pendingFieldUpdate struct {
	ctx        contextx.IContext
	model      mongo.WriteModel
	mu         sync.Mutex
	dispatched bool // Protected by mu, together with the cancellation decision.
}

type updateFieldBatcher struct {
	dao       IDao
	immediate updateExecutor
	combined  combine.IHandler[*pendingFieldUpdate, []error]
}

func newUpdateFieldBatcher(dao IDao, immediate updateExecutor) *updateFieldBatcher {
	batcher := &updateFieldBatcher{dao: dao, immediate: immediate}
	batcher.combined = combine.New(updateFieldBatchLimit, updateFieldBatchGap, batcher.send)

	return batcher
}

func (batcher *updateFieldBatcher) UpdateMany(
	nCtx contextx.IContext, filter bson.D, update bson.D,
) error {

	if mongo.SessionFromContext(nCtx) != nil {
		return batcher.immediate.UpdateMany(nCtx, filter, update)
	}
	tracer := trace.SpanFromContext(nCtx).TracerProvider().Tracer("dao.mongo.base")
	ctx, span := tracer.Start(nCtx, "mongo.batch.update_many", trace.WithAttributes(
		attribute.String(attrKeyMongoCollection, batcher.dao.GetTableName()),
	))
	defer span.End()
	err := batcher.updateMany(contextx.FromContext(ctx), filter, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (batcher *updateFieldBatcher) updateMany(
	nCtx contextx.IContext, filter bson.D, update bson.D,
) error {

	if err := nCtx.Err(); err != nil {
		return err
	}
	// Encode before admission so a caller's invalid value cannot fail its peers.
	encodedFilter, err := bson.Marshal(filter)
	if err != nil {
		return fmt.Errorf("encode batched update filter: %w", err)
	}
	encodedUpdate, err := bson.Marshal(update)
	if err != nil {
		return fmt.Errorf("encode batched update fields: %w", err)
	}
	item := &pendingFieldUpdate{
		ctx: nCtx,
		model: mongo.NewUpdateManyModel().SetFilter(bson.Raw(encodedFilter)).
			SetUpdate(bson.Raw(encodedUpdate)).SetUpsert(false),
	}
	// combine owns admission and result delivery. Only queued calls can abandon
	// waiting; once dispatched, a caller must observe its actual write result.
	waitCtx, cancel := context.WithCancel(contextx.WithoutCancel(nCtx))
	defer cancel()
	stop := context.AfterFunc(nCtx, func() {
		item.mu.Lock()
		defer item.mu.Unlock()
		if !item.dispatched {
			cancel()
		}
	})
	defer stop()
	results, index, err := batcher.combined.Call(waitCtx, item)
	if err != nil {
		if ctxErr := nCtx.Err(); ctxErr != nil {
			return ctxErr
		}

		return err
	}

	return results[index]
}

func (batcher *updateFieldBatcher) send(_ string, batch []*pendingFieldUpdate) ([]error, error) {
	results := make([]error, len(batch))
	for offset := 0; offset < len(batch); offset += updateFieldBatchLimit {
		end := min(offset+updateFieldBatchLimit, len(batch))
		copy(results[offset:end], batcher.flush(batch[offset:end]))
	}

	return results, nil
}

func (batcher *updateFieldBatcher) flush(batch []*pendingFieldUpdate) []error {
	results := make([]error, len(batch))
	items := make([]*pendingFieldUpdate, 0, len(batch))
	indexes := make([]int, 0, len(batch))
	models := make([]mongo.WriteModel, 0, len(batch))
	links := make([]trace.Link, 0, len(batch))
	for index, item := range batch {
		item.mu.Lock()
		err := item.ctx.Err()
		item.dispatched = true
		item.mu.Unlock()
		if err != nil {
			results[index] = err

			continue
		}
		items = append(items, item)
		indexes = append(indexes, index)
		models = append(models, item.model)
		span := trace.SpanFromContext(item.ctx)
		links = append(links, trace.Link{SpanContext: span.SpanContext()})
	}
	if len(items) == 0 {
		return results
	}

	nCtx, cancel := contextx.WithTimeout(contextx.WithoutCancel(items[0].ctx), updateFieldBatchTimeout)
	defer cancel()
	tracer := trace.SpanFromContext(nCtx).TracerProvider().Tracer("dao.mongo.base")
	ctx, span := tracer.Start(nCtx, "mongo.batch.flush", trace.WithLinks(links...), trace.WithAttributes(
		attribute.String(attrKeyMongoCollection, batcher.dao.GetTableName()),
		attribute.Int(attrKeyBatchSize, len(models)),
	))
	nCtx = contextx.FromContext(ctx)
	metric := newMetricData(batcher.dao.GetTableName()).start(daomongo.MetricOperationBulkWrite, len(models))
	result, err := batcher.dao.GetClient().BulkWrite(nCtx, models, mongoOptions.BulkWrite().SetOrdered(false))
	modified := 0
	if result != nil {
		modified = int(result.ModifiedCount)
	}
	metric.end(err, modified)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
	for index, item := range items {
		trace.SpanFromContext(item.ctx).SetAttributes(attribute.Int(attrKeyBatchSize, len(models)))
		results[indexes[index]] = fieldUpdateBatchError(err, index)
	}

	return results
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
