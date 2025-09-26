/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package stopoperinst define the table to store stopping operation instance.
package stopoperinst

import (
	"context"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database) *dao {
	return &dao{client: client.Collection(TableName)}
}

// buildTTLIndexModel build ttl index model, this index is used to delete expired data.
func buildTTLIndexModel() mongo.IndexModel {
	return mongo.IndexModel{
		Keys: bson.D{{Key: "data.expire_at", Value: 1}},
		Options: mongoOptions.Index().
			SetExpireAfterSeconds(0),
	}
}

type dao struct {
	client *mongo.Collection
}

// upsert updates or inserts an operation instance data.
func (d *dao) upsert(nCtx contextx.IContext, inst *StopOperInst) error {
	indexModel := buildTTLIndexModel()
	if _, err := d.client.Indexes().CreateOne(nCtx, indexModel); err != nil {
		logger.G.Sys().WithErr(err).Error("failed to create ttl index")

		return err
	}

	filter, upsert, opts := buildUpsertParams(inst)
	result, err := d.client.UpdateOne(nCtx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		logger.G.Sys().With("unique-key", inst.UniqueKey()).Info("inserted stop operation instance")

	case result.MatchedCount > 0:
		logger.G.Sys().With("unique-key", inst.UniqueKey()).Info("updated stop operation instance")

	default:
		logger.G.Sys().With("unique-key", inst.UniqueKey()).Info("try to upsert stop operation instance but no changes made")
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(inst *StopOperInst) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update data by operation_inst_data_id.
	filter := bson.D{{Key: "data.oper_inst_id", Value: inst.OperInstID}}

	// upsert as creation or update data only.
	update := base.BuildUpsertParam(inst)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// find all stopping operation instance.
func (d *dao) find(nCtx contextx.IContext, filter bson.D) ([]*StopOperInst, error) {
	result, err := d.client.Find(nCtx, filter)
	if err != nil {
		return nil, err
	}

	stopOperInsts := make([]*StopOperInst, 0)
	for result.Next(nCtx) {
		table := &TableStopOperInst{}
		if err := result.Decode(table); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to decode stopping operation instance")

			continue
		}
		stopOperInsts = append(stopOperInsts, table.Data)
	}

	return stopOperInsts, nil
}

// watchWithRetry watch with retry.
func (d *dao) watch(nCtx contextx.IContext, filter bson.D, fn func(*StopOperInst)) error {
	pipeline, watchOptions := buildWatchParams(filter)
	changeStream, err := d.client.Watch(context.Background(), pipeline, watchOptions)
	if err != nil {
		logger.G.Sys().WithErr(err).Error("failed to watch stopping operation instances")

		return err
	}

	defer changeStream.Close(nCtx)

	for {
		select {
		case <-nCtx.Done():
			return nil
		default:
			if changeStream.Next(nCtx) {
				changeEvent := &TableStopOperInstChangeEvent{}
				if err := changeStream.Decode(changeEvent); err != nil {
					logger.G.Sys().WithErr(err).Error("failed to decode change event")

					continue
				}

				if changeEvent.FullDocument == nil {
					logger.G.Sys().Warn("full document is nil, skip")

					continue
				}

				fn(&StopOperInst{
					OperInstID: changeEvent.FullDocument.Data.OperInstID,
					ExpireAt:   changeEvent.FullDocument.Data.ExpireAt,
				})
			}

			if err := changeStream.Err(); err != nil {
				logger.G.Sys().WithErr(err).Error("failed to watch stopping operation instances")

				return err
			}
		}
	}
}

// watchWithRetry watch with retry.
func (d *dao) watchWithRetry(nCtx contextx.IContext, filter bson.D, fn func(*StopOperInst)) {
	backoff := time.Second
	maxBackoff := time.Minute

	logger.G.Sys().Info("start to watch stopping operation instances")
	for {
		err := d.watch(nCtx, filter, fn)
		if err == nil {
			return
		}

		logger.G.Sys().WithErr(err).With("retry-after", backoff).Warn("failed to watch stopping operation instances")
		select {
		case <-nCtx.Done():
			return
		case <-time.After(backoff):
			// exponential backoff
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// buildUpsertParams build update params.
func buildWatchParams(filter bson.D) (mongo.Pipeline, *mongoOptions.ChangeStreamOptions) {
	// only watch insert event
	pipeline := mongo.Pipeline{{{Key: "$match", Value: filter}}}

	watchOptions := mongoOptions.ChangeStream().
		SetBatchSize(1000).
		SetMaxAwaitTime(time.Second * 10)

	return pipeline, watchOptions
}
