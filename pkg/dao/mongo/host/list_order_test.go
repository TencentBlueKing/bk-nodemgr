/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package host

import (
	"reflect"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestDaoGetIndexes_HostListSort(t *testing.T) {
	indexes := (&dao{}).GetIndexes()
	wantIndexes := []bson.D{
		{
			{Key: FieldKeyStaticBizID, Value: 1},
			{Key: FieldKeyDynamicNodeRole, Value: 1},
			{Key: FieldKeyOperationUpdatedAt, Value: -1},
		},
		{
			{Key: FieldKeyStaticSetID, Value: 1},
			{Key: FieldKeyDynamicNodeRole, Value: 1},
			{Key: FieldKeyOperationUpdatedAt, Value: -1},
		},
		{
			{Key: FieldKeyStaticModuleID, Value: 1},
			{Key: FieldKeyDynamicNodeRole, Value: 1},
			{Key: FieldKeyOperationUpdatedAt, Value: -1},
		},
		{
			{Key: FieldKeyDynamicNodeRole, Value: 1},
			{Key: FieldKeyOperationUpdatedAt, Value: -1},
		},
	}

	for _, wantIndex := range wantIndexes {
		if !hasPartialIndex(indexes, wantIndex) {
			t.Fatalf("GetIndexes() missing partial index %v", wantIndex)
		}
	}

	legacyIndex := bson.D{
		{Key: FieldKeyStaticBizID, Value: 1},
		{Key: FieldKeyDynamicNodeRole, Value: 1},
		{Key: base.FieldKeyUpdatedAt, Value: -1},
	}
	if hasPartialIndex(indexes, legacyIndex) {
		t.Fatalf("GetIndexes() contains legacy partial index %v", legacyIndex)
	}
}

func hasPartialIndex(indexes []mongo.IndexModel, keys bson.D) bool {
	for _, index := range indexes {
		if !reflect.DeepEqual(index.Keys, keys) {
			continue
		}
		if hasAlivePartialFilter(index.Options) {
			return true
		}
	}

	return false
}

func hasAlivePartialFilter(indexOptions *options.IndexOptions) bool {
	if indexOptions == nil || indexOptions.PartialFilterExpression == nil {
		return false
	}

	wantFilter := bson.D{{Key: base.FieldKeyIsDeleted, Value: false}}
	return reflect.DeepEqual(indexOptions.PartialFilterExpression, wantFilter)
}
