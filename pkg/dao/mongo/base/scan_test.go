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
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestOrm_ScanWithLimit(t *testing.T) {
	tests := []struct {
		name        string
		limit       int64
		batches     []int
		queryLimits []int64
	}{
		{name: "below batch size", limit: 499, batches: []int{499}, queryLimits: []int64{499}},
		{name: "exact batch size", limit: 500, batches: []int{500}, queryLimits: []int64{500}},
		{name: "partial final batch", limit: 501, batches: []int{500, 1}, queryLimits: []int64{500, 1}},
		{name: "exact multiple", limit: 1000, batches: []int{500, 500}, queryLimits: []int64{500, 500}},
		{name: "fewer matches", limit: 1000, batches: []int{500, 1}, queryLimits: []int64{500, 500}},
		{name: "no matches", limit: 1000, batches: []int{0}, queryLimits: []int64{500}},
		{name: "existing unlimited scan", limit: 0, batches: []int{500, 1}, queryLimits: []int64{500, 500}},
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, tt := range tests {
		mt.Run(tt.name, func(mt *mtest.T) {
			total := 0
			for _, count := range tt.batches {
				documents := make([]bson.D, count)
				for i := range count {
					id := total + i + 1
					documents[i] = bson.D{
						{Key: "_id", Value: int64(id)},
						{Key: "data", Value: bson.D{{Key: "id", Value: strconv.Itoa(id)}}},
					}
				}
				total += count
				mt.AddMockResponses(mtest.CreateCursorResponse(0, mt.DB.Name()+"."+mt.Coll.Name(), mtest.FirstBatch, documents...))
			}
			orm := NewOrm[*TestData, TestData](&TestDao{collection: mt.Coll, tableName: mt.Coll.Name()})
			nCtx := contextx.New(context.Background())
			filter := AliveFilter()
			var data []*TestData
			var err error
			if tt.limit == 0 {
				data, err = orm.ScanAll(nCtx, filter)
			} else {
				data, err = orm.ScanWithLimit(nCtx, filter, tt.limit)
			}
			require.NoError(mt, err)
			require.Len(mt, data, total)
			for i, item := range data {
				require.Equal(mt, strconv.Itoa(i+1), item.ID)
			}
			require.Equal(mt, AliveFilter(), filter)
			for i, limit := range tt.queryLimits {
				event := mt.GetStartedEvent()
				require.NotNil(mt, event)
				require.Equal(mt, "find", event.CommandName)
				require.Equal(mt, limit, event.Command.Lookup("limit").Int64())
				require.Equal(mt, int32(1), event.Command.Lookup("sort").Document().Lookup("_id").Int32())
				if i == 0 {
					continue
				}
				conditions, err := event.Command.Lookup("filter").Document().Lookup("$and").Array().Values()
				require.NoError(mt, err)
				cursor := conditions[len(conditions)-1].Document().Lookup("_id").Document().Lookup("$gt").Int64()
				require.Equal(mt, int64(i*500), cursor)
			}
			require.Nil(mt, mt.GetStartedEvent(), "scan must stop without another query after reaching the limit")
		})
	}
}

func TestOrm_ScanWithLimitRejectsNonPositiveLimit(t *testing.T) {
	orm := new(Orm[*TestData, TestData])
	for _, limit := range []int64{0, -1} {
		data, err := orm.ScanWithLimit(contextx.New(context.Background()), nil, limit)
		require.Error(t, err)
		require.Nil(t, data)
	}
}
