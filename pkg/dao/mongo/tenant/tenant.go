/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package tenant ...
package tenant

import (
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

type dao struct {
	client *mongo.Collection
}

// upsert updates or inserts a tenant.
func (d *dao) upsert(nCtx contextx.IContext, tenant *Tenant) error {
	filter, upsert, opts := buildUpsertParams(tenant)
	result, err := d.client.UpdateOne(nCtx, filter, upsert, opts)
	if err != nil {
		return err
	}

	switch {
	case result.UpsertedCount > 0:
		logger.G.Sys().With("unique-key", tenant.UniqueKey()).Info("inserted tenant")

	case result.MatchedCount > 0:
		logger.G.Sys().With("unique-key", tenant.UniqueKey()).Info("updated tenant")

	default:
		logger.G.Sys().With("unique-key", tenant.UniqueKey()).Info("try to upsert tenant but no changes made")
	}

	return nil
}

// buildUpsertParams build update params.
func buildUpsertParams(tenant *Tenant) (bson.D, bson.D, *mongoOptions.UpdateOptions) {
	// update tenant by tenant_id.
	filter := bson.D{{Key: "data.tenant_id", Value: tenant.ID}}

	// insert as creation or update data only.
	update := base.BuildUpsertParam(tenant)

	// do upsert.
	opts := mongoOptions.Update().SetUpsert(true)

	return filter, update, opts
}

// ListAll list all tenant.
func (d *dao) listAll(nCtx contextx.IContext) ([]*Tenant, error) {
	result, err := d.client.Find(nCtx, bson.D{{Key: "basic.is_deleted", Value: false}})
	if err != nil {
		return nil, err
	}

	tenants := make([]*Tenant, 0)
	for result.Next(nCtx) {
		table := &TableTenant{}
		if err := result.Decode(table); err != nil {
			logger.G.Sys().WithErr(err).Warn("failed to decode tenant")

			continue
		}
		tenants = append(tenants, table.Data)
	}

	return tenants, nil
}
