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

// Package deploypolicyworkflow stores tenant-scoped deploy policy execution records.
package deploypolicyworkflow

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

func newDao(client *mongo.Database, tenantID string) *dao {
	d := &dao{client: client.Collection(TableName(tenantID)), tableName: TableName(tenantID)}
	d.IOrm = base.NewOrm[*Data, Data](d)

	return d
}

type dao struct {
	client    *mongo.Collection
	tableName string
	base.IOrm[*Data, Data]
}

// GetClient returns the tenant collection.
func (d *dao) GetClient() *mongo.Collection { return d.client }

// GetTableName returns the tenant collection name.
func (d *dao) GetTableName() string { return d.tableName }

// GetIndexes adds operation and policy idempotency to the ORM's workflow identity index.
func (d *dao) GetIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{{
		Keys: bson.D{
			{Key: FieldKeyOperationID, Value: 1},
			{Key: FieldKeyDeployPolicyID, Value: 1},
		},
		Options: mongoOptions.Index().SetUnique(true),
	}}
}

func (d *dao) recordChild(nCtx contextx.IContext, workflowID string, child Child) error {
	identity := bson.M{"workflow_id": child.WorkflowID, "workflow_domain": child.WorkflowDomain}
	absent := bson.D{
		{Key: base.FieldKeyIsDeleted, Value: false},
		{Key: FieldKeyWorkflowID, Value: workflowID},
		{Key: FieldKeyChildren, Value: bson.M{"$not": bson.M{"$elemMatch": identity}}},
	}
	result, err := d.client.UpdateOne(nCtx, absent, bson.M{
		"$push": bson.M{FieldKeyChildren: child},
		"$set":  bson.M{base.FieldKeyUpdatedAt: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("failed to append child to workflow %s: %w", workflowID, err)
	}
	if result.MatchedCount == 1 {
		return nil
	}

	// Another writer may have appended the same child. Only promote confirmation, never replace it.
	present := bson.D{
		{Key: base.FieldKeyIsDeleted, Value: false},
		{Key: FieldKeyWorkflowID, Value: workflowID},
		{Key: FieldKeyChildren, Value: bson.M{"$elemMatch": identity}},
	}
	fields := bson.M{base.FieldKeyUpdatedAt: time.Now()}
	if child.Confirmed {
		fields[FieldKeyChildren+".$.confirmed"] = true
	}
	result, err = d.client.UpdateOne(nCtx, present, bson.M{"$set": fields})
	if err != nil {
		return fmt.Errorf("failed to acknowledge child for workflow %s: %w", workflowID, err)
	}
	if result.MatchedCount != 1 {
		return fmt.Errorf("workflow %s not found: %w", workflowID, base.ErrRecordNoFound())
	}

	return nil
}
