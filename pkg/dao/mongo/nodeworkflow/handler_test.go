/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeworkflow

import (
	"context"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/counter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/test/mongodaotest"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestSuite ...
type TestSuite struct {
	mongodaotest.TestSuite[*Data, Data]
	Handler IHandler
	counter counter.Handler
}

// TestAll is the entry point for all tests in this package.
func TestAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuit := new(TestSuite)
	testSuit.TestSuite = mongodaotest.NewMongoDaoTestSuite[*Data, Data](logger.LoggerDefault{}, func(client *mongo.Database, logger logger.Logger) {
		testSuit.Dao = newDao("", client, logger)
		testSuit.Handler = New(client, logger)
		testSuit.counter = counter.New(client, logger)
		testSuit.TestDatas = testSuit.prepareTestData()
	})

	suite.Run(t, testSuit)
}

func (testSuit *TestSuite) prepareTestData() []*Data {
	testDatas := []*Data{
		{
			WorkflowID: "1",
			TriggerID:  "T-123456",
			Type:       string(types.NodeWorkflowTypeInstallAgent),
			BizIDs:     []int64{639},
			Status:     "running",
		},
		{
			WorkflowID: "1",
			TriggerID:  "T-123457",
			Type:       string(types.NodeWorkflowTypeInstallAgent),
			BizIDs:     []int64{639},
			Status:     "failed",
		},
		{
			WorkflowID: "2",
			TriggerID:  "T-123458",
			Type:       "uninstall",
			BizIDs:     []int64{639},
			Status:     "running",
		},
		{
			WorkflowID: "2",
			TriggerID:  "T-123459",
			Type:       "uninstall",
			BizIDs:     []int64{639},
			Status:     "failed",
		},
	}

	return testDatas
}

// TestCreate tests the Create method of the handler.
func (testSuit *TestSuite) TestCreate() {
	type args struct {
		ctx          context.Context
		nodeWorkflow *types.NodeWorkflow
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx: context.Background(),
				nodeWorkflow: &types.NodeWorkflow{
					WorkflowID: "1",
					TriggerID:  "T-123459",
					Type:       "install",
					Status:     types.NodeWorkflowStatusRunning,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			err := testSuit.Handler.Create(tt.args.ctx, tt.args.nodeWorkflow)
			if !tt.wantErr {
				testSuit.NoErrorf(err, "Create() error = %v", err)
			}

			testSuit.T().Logf("err: %v", err)
		})
	}
}

// TestList tests the List method of the handler.
func (testSuit *TestSuite) TestList() {
	type args struct {
		ctx  context.Context
		page types.Page
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    []*types.NodeWorkflow
		wantNum int64
		wantErr bool
	}{
		{
			name: "filter by workflow id",
			args: args{
				ctx: context.Background(),
				page: types.Page{
					Offset: 0,
					Limit:  1,
					Sort:   "",
				},
				opts: []OptFn{
					WithWorkflowID("1"),
				},
			},
			want: []*types.NodeWorkflow{
				{
					WorkflowID: "2",
					TriggerID:  "T-123457",
					Type:       "install",
					BizIDs:     []int64{639},
					Status:     "failed",
				},
			},
			wantNum: 1,
			wantErr: false,
		},
		{
			name: "filter by status",
			args: args{
				ctx: context.Background(),
				page: types.Page{
					Offset: 0,
					Limit:  2,
					Sort:   "",
				},
				opts: []OptFn{
					WithStatus(types.NodeWorkflowStatusRunning),
				},
			},
			want: []*types.NodeWorkflow{
				{
					WorkflowID: "0",
					TriggerID:  "T-123456",
					Type:       "install",
					BizIDs:     []int64{639},

					Status: "running",
				},
				{
					WorkflowID: "0",
					TriggerID:  "T-123458",
					Type:       "uninstall",

					Status: "running",
				},
			},
			wantNum: 2,
			wantErr: false,
		},
		{
			name: "filter by oper type",
			args: args{
				ctx: context.Background(),
				page: types.Page{
					Offset: 0,
					Limit:  1,
				},
				opts: []OptFn{
					WithWorkflowID("1"),
					WithType(types.NodeWorkflowTypeInstallAgent),
				},
			},
			want: []*types.NodeWorkflow{
				{
					WorkflowID: "1",
					TriggerID:  "T-123457",
					Type:       types.NodeWorkflowTypeInstallAgent,
					BizIDs:     []int64{639},

					Status: "failed",
				},
			},
			wantNum: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			got, gotNum, err := testSuit.Handler.List(tt.args.ctx, tt.args.page, tt.args.opts...)
			if !tt.wantErr {
				testSuit.NoErrorf(err, "List() error = %v", err)
			}

			testSuit.Equal(tt.wantNum, gotNum)
			testSuit.Equal(tt.want, got)
		})
	}
}

// TestCount tests the Count method of the handler.
func (testSuit *TestSuite) TestCount() {
	type args struct {
		ctx  context.Context
		opts []OptFn
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "filter by workflow id",
			args: args{
				ctx: context.Background(),
				opts: []OptFn{
					WithWorkflowID("1"),
				},
			},
			want:    1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			got, err := testSuit.Handler.Count(tt.args.ctx, tt.args.opts...)
			if !tt.wantErr {
				testSuit.NoErrorf(err, "Count() error = %v", err)
			}

			testSuit.Equal(tt.want, got)
		})
	}
}

// TestUpdateStatus tests the UpdateStatus method of the handler.
func (testSuit *TestSuite) TestUpdateStatus() {
	type args struct {
		ctx        context.Context
		workflowID string
		status     types.NodeWorkflowStatus
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:        context.Background(),
				workflowID: "1",
				status:     types.NodeWorkflowStatusSuccess,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			err := testSuit.Handler.UpdateStatus(tt.args.ctx, tt.args.workflowID, tt.args.status)
			if !tt.wantErr {
				testSuit.NoErrorf(err, "UpdateStatus() error = %v", err)
			}

			testSuit.T().Logf("err: %v", err)
		})
	}
}
