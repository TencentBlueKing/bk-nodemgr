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

// Package base provides base orm for mongodb.
// nolint: nonamedreturns
package base

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daomongo "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const scanBatchSize int64 = 500

type scanDocument[P IData] struct {
	ID        any       `bson:"_id"`
	BasicInfo BasicInfo `json:"basic" bson:"basic"`
	Data      P         `json:"data" bson:"data"`
}

// ScanAll scans all data by given filter.
func (orm *Orm[P, T]) ScanAll(nCtx contextx.IContext, filter bson.D, field ...string) (dataPoints []P, err error) {
	metric := orm.metric().start(daomongo.MetricOperationScanAll, len(filter))
	defer func() {
		metric.end(err, len(dataPoints))

		duration := time.Since(metric.startTime)
		if duration < daomongo.DefaultSlowTime || nCtx == nil {
			return
		}

		span := trace.SpanFromContext(nCtx)
		if !span.SpanContext().IsValid() {
			return
		}

		span.AddEvent(spanEventSlowQuery, trace.WithAttributes(
			attribute.String(attrKeyORMCollection, orm.dao.GetTableName()),
			attribute.String(attrKeyORMOperation, "scan_all"),
			attribute.Int64(attrKeyORMDurationMS, duration.Milliseconds()),
			attribute.Int(attrKeyORMFilterSize, len(filter)),
			attribute.Int(attrKeyORMResultCount, len(dataPoints)),
		))
	}()

	return orm.scan(nCtx, filter, 0, field...)
}

// ScanWithLimit scans at most limit documents by given filter. The limit must be positive.
func (orm *Orm[P, T]) ScanWithLimit(
	nCtx contextx.IContext, filter bson.D, limit int64, field ...string) (dataPoints []P, err error) {

	metric := orm.metric().start(daomongo.MetricOperationScanWithLimit, len(filter))
	defer func() {
		metric.end(err, len(dataPoints))

		duration := time.Since(metric.startTime)
		if duration < daomongo.DefaultSlowTime || nCtx == nil {
			return
		}

		span := trace.SpanFromContext(nCtx)
		if !span.SpanContext().IsValid() {
			return
		}

		span.AddEvent(spanEventSlowQuery, trace.WithAttributes(
			attribute.String(attrKeyORMCollection, orm.dao.GetTableName()),
			attribute.String(attrKeyORMOperation, "scan_with_limit"),
			attribute.Int64(attrKeyORMDurationMS, duration.Milliseconds()),
			attribute.Int(attrKeyORMFilterSize, len(filter)),
			attribute.Int(attrKeyORMResultCount, len(dataPoints)),
		))
	}()

	if limit <= 0 {
		return nil, errors.New("scan limit must be positive")
	}

	return orm.scan(nCtx, filter, limit, field...)
}

// scan scans matching data in cursor batches. A zero limit returns all matches; a positive limit caps the result count.
func (orm *Orm[P, T]) scan(nCtx contextx.IContext, filter bson.D, limit int64, field ...string) (dataPoints []P, err error) {
	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	findOpt := buildScanFindOptions(field)
	dataPoints = make([]P, 0)

	var lastID any
	for {
		batchLimit := scanBatchSize
		if limit > 0 {
			batchLimit = min(batchLimit, limit-int64(len(dataPoints)))
		}
		findOpt.SetLimit(batchLimit)
		queryFilter := buildScanFilter(filter, lastID)
		cursor, err := orm.dao.GetClient().Find(nCtx, queryFilter, findOpt)
		if err != nil {
			return nil, err
		}

		batchCount, advanced, scanErr := orm.scanBatch(nCtx, cursor, &lastID, &dataPoints)
		if scanErr != nil {
			return nil, scanErr
		}
		if batchCount < batchLimit || (limit > 0 && int64(len(dataPoints)) >= limit) {
			break
		}
		if !advanced {
			return nil, errors.New("failed to scan documents: cursor did not advance")
		}
	}

	return dataPoints, nil
}

func buildScanFindOptions(fields []string) *mongoOptions.FindOptions {
	findOpt := mongoOptions.Find().SetSort(bson.D{{Key: "_id", Value: 1}}).SetLimit(scanBatchSize)
	if len(fields) == 0 {
		return findOpt
	}

	projection := make(bson.D, 0, len(fields))
	for _, field := range fields {
		projection = append(projection, bson.E{Key: field, Value: 1})
	}
	findOpt.SetProjection(projection)

	return findOpt
}

func buildScanFilter(filter bson.D, lastID any) bson.D {
	if lastID == nil {
		return filter
	}

	condition := bson.D{{Key: "_id", Value: bson.D{{Key: "$gt", Value: lastID}}}}

	return appendAndCondition(cloneScanFilter(filter), condition)
}

func cloneScanFilter(filter bson.D) bson.D {
	cloned := make(bson.D, len(filter))
	copy(cloned, filter)
	for i := range cloned {
		if cloned[i].Key != "$and" {
			continue
		}

		conditions, ok := cloned[i].Value.(bson.A)
		if !ok {
			continue
		}

		clonedConditions := make(bson.A, len(conditions))
		copy(clonedConditions, conditions)
		cloned[i].Value = clonedConditions
	}

	return cloned
}

func (orm *Orm[P, T]) scanBatch(
	nCtx contextx.IContext, cursor *mongo.Cursor, lastID *any, dataPoints *[]P) (batchCount int64, advanced bool, err error) {

	defer func() {
		if closeErr := cursor.Close(nCtx); closeErr != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(closeErr).With("table", orm.dao.GetTableName()).Warn("failed to close scan cursor")
		}
	}()

	for cursor.Next(nCtx) {
		batchCount++

		document := &scanDocument[P]{}
		if err := cursor.Decode(document); err != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(err).With("table", orm.dao.GetTableName()).Info("failed to scan, failed to decode document")

			continue
		}
		if document.ID == nil {
			return batchCount, advanced, errors.New("failed to scan documents: missing _id")
		}

		*dataPoints = append(*dataPoints, document.Data)
		*lastID = document.ID
		advanced = true
	}

	if err := cursor.Err(); err != nil {
		return batchCount, advanced, fmt.Errorf("failed to scan documents: %w", err)
	}

	return batchCount, advanced, nil
}
