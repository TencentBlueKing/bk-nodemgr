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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	daomongo "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AggregateDistinctField describes one field selected for aggregate distinct.
type AggregateDistinctField struct {
	ResultKey string
	FieldPath string
	BSONType  bsontype.Type
}

// AggregateDistinctResult contains raw distinct values keyed by result key.
type AggregateDistinctResult map[string][]bson.RawValue

// AggregateDistinct returns distinct values for multiple fields in one aggregation.
//
//nolint:gocognit // Keep slow-query tracing attributes explicit at the operation site.
func AggregateDistinct(
	nCtx contextx.IContext, dao IDao, filter bson.D, fields []AggregateDistinctField,
) (AggregateDistinctResult, error) {

	result := newAggregateDistinctResult(fields)
	if len(fields) == 0 {
		return result, nil
	}
	if nCtx == nil {
		return nil, errors.New("context is nil")
	}

	var aggregateErr error
	metric := newMetricData(dao.GetTableName()).start(daomongo.MetricOperationDistinct, len(filter))
	defer func() {
		metric.end(aggregateErr, aggregateDistinctResultCount(result))

		duration := time.Since(metric.startTime)
		if duration < daomongo.DefaultSlowTime {
			return
		}

		span := trace.SpanFromContext(nCtx)
		if !span.SpanContext().IsValid() {
			return
		}

		filterJSON, err := bson.MarshalExtJSON(filter, false, false)
		if err != nil {
			logger.G.Sys().Ctx(nCtx).WithErr(err).
				With("table", dao.GetTableName(), "operation", "aggregate_distinct").
				Warn("failed to marshal slow query filter")
		}

		var filterAttrs []attribute.KeyValue
		if len(filterJSON) > 0 {
			filterAttrs = []attribute.KeyValue{attribute.String(attrKeyORMFilter, string(filterJSON))}
		}

		span.AddEvent(spanEventSlowQuery,
			trace.WithAttributes(
				attribute.String(attrKeyORMCollection, dao.GetTableName()),
				attribute.String(attrKeyORMOperation, "aggregate_distinct"),
				attribute.Int64(attrKeyORMDurationMS, duration.Milliseconds()),
				attribute.Int(attrKeyORMFilterSize, len(filter)),
				attribute.Int(attrKeyORMResultCount, aggregateDistinctResultCount(result)),
			),
			trace.WithAttributes(filterAttrs...),
		)
	}()

	cursor, err := dao.GetClient().Aggregate(
		nCtx,
		buildAggregateDistinctPipeline(filter, fields),
		aggregateDistinctOptions(),
	)
	if err != nil {
		aggregateErr = err

		return nil, fmt.Errorf("failed to aggregate distinct: %w", err)
	}
	defer func() {
		if closeErr := cursor.Close(nCtx); closeErr != nil {
			logger.G.Sys().WithErr(closeErr).
				With("table", dao.GetTableName(), "filter-size", len(filter)).
				Error("failed to close cursor of aggregate distinct")
		}
	}()

	var document bson.Raw
	if cursor.Next(nCtx) {
		if err := cursor.Decode(&document); err != nil {
			aggregateErr = err

			return nil, fmt.Errorf("failed to decode aggregate distinct result: %w", err)
		}

		decodedResult, err := decodeAggregateDistinctResult(document, fields)
		if err != nil {
			aggregateErr = err

			return nil, err
		}
		result = decodedResult
	}

	if err := cursor.Err(); err != nil {
		aggregateErr = err

		return nil, fmt.Errorf("failed to iterate aggregate distinct result: %w", err)
	}
	if cursor.Next(nCtx) {
		aggregateErr = errors.New("aggregate distinct returned more than one document")

		return nil, aggregateErr
	}
	if err := cursor.Err(); err != nil {
		aggregateErr = err

		return nil, fmt.Errorf("failed to iterate aggregate distinct result: %w", err)
	}

	return result, nil
}

func buildAggregateDistinctPipeline(filter bson.D, fields []AggregateDistinctField) mongo.Pipeline {
	group := make(bson.D, 1, len(fields)+1)
	group[0] = bson.E{Key: "_id", Value: nil}
	for _, field := range fields {
		group = append(group, bson.E{
			Key: field.ResultKey,
			Value: bson.D{{
				Key:   "$addToSet",
				Value: "$" + field.FieldPath,
			}},
		})
	}

	return mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: group}},
	}
}

func aggregateDistinctOptions() *mongoOptions.AggregateOptions {
	return mongoOptions.Aggregate().SetAllowDiskUse(true)
}

func newAggregateDistinctResult(fields []AggregateDistinctField) AggregateDistinctResult {
	result := make(AggregateDistinctResult, len(fields))
	for _, field := range fields {
		result[field.ResultKey] = make([]bson.RawValue, 0)
	}

	return result
}

func decodeAggregateDistinctResult(document bson.Raw, fields []AggregateDistinctField) (AggregateDistinctResult, error) {
	result := newAggregateDistinctResult(fields)
	for _, field := range fields {
		value := document.Lookup(field.ResultKey)
		switch value.Type {
		case bsontype.Type(0), bson.TypeNull:
			continue
		case bson.TypeArray:
		default:
			return nil, fmt.Errorf("aggregate distinct result field(%s) is not an array", field.ResultKey)
		}

		values, err := value.Array().Values()
		if err != nil {
			return nil, fmt.Errorf("failed to decode aggregate distinct result field(%s): %w", field.ResultKey, err)
		}
		for _, rawValue := range values {
			if rawValue.Type == field.BSONType {
				result[field.ResultKey] = append(result[field.ResultKey], rawValue)
			}
		}
	}

	return result, nil
}

func aggregateDistinctResultCount(result AggregateDistinctResult) int {
	count := 0
	for _, values := range result {
		count += len(values)
	}

	return count
}
