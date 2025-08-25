/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodedeployment

import (
	"context"
	"testing"

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
}

// TestAll is the entry point for all tests in this package.
func TestAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testSuit := new(TestSuite)
	testSuit.TestSuite = mongodaotest.NewMongoDaoTestSuite[*Data, Data](logger.LoggerDefault{}, func(client *mongo.Database, logger logger.ILogger) {
		testSuit.Dao = newDao(client, logger)
		testSuit.Handler = New(client, logger)
		testSuit.TestDatas = prepareTestData()
	})

	suite.Run(t, testSuit)
}

func prepareTestData() []*Data {
	testDatas := []*Data{
		{
			Token: "token123456",
			Info: &Info{
				OperInstID:       "inst-001",
				ActionName:       "wait_agent_install",
				HostID:           52296,
				OSType:           "linux",
				TenantID:         "single",
				NodeRole:         "agent",
				NodeStatus:       "running",
				NodeVersion:      "v2.1.6-beta.55",
				NodeGeneration:   2,
				AgentID:          "02000000005254001bbe721742528553406b",
				NetworkUnitID:    0,
				NetworkAreaID:    0,
				BizID:            12,
				InnerIP:          "192.168.1.10",
				Addressing:       "static",
				ProxyClusterPort: 0,
				ProxyDataPort:    0,
				ProxyFilePort:    0,
			},
			NodeConf: &NodeConf{
				PreSetting: map[string]any{
					"maxConnections": int64(1000),
					"timeout":        int64(30),
					"retries":        int64(3),
				},
				CustomSetting: map[string]any{
					"city_id":  "info",
					"run_mode": "agent",
					"cloud_id": int64(0),
					"zone_id":  "1",
				},
			},
		},
		{
			Token: "token789012",
			Info: &Info{
				OperInstID:     "inst-002",
				ActionName:     "update",
				HostID:         67890,
				OSType:         "windows",
				TenantID:       "tenant-def",
				NodeRole:       "agent",
				NodeStatus:     "initializing",
				NodeVersion:    "v1.0.2",
				NodeGeneration: 2,
				AgentID:        "agent-002",
				NetworkUnitID:  101,
				NetworkAreaID:  201,
				BizID:          301,
				InnerIP:        "192.168.1.11",
				Addressing:     "dhcp",
			},
			NodeConf: &NodeConf{
				PreSetting: map[string]any{
					"maxConnections": 500,
					"timeout":        60,
					"bufferSize":     4096,
				},
				CustomSetting: map[string]any{
					"logLevel":       "debug",
					"enableMetrics":  false,
					"securityPolicy": "strict",
				},
			},
		},
		{
			Token: "token345678",
			Info: &Info{
				OperInstID:     "inst-003",
				ActionName:     "restart",
				HostID:         24680,
				OSType:         "macos",
				TenantID:       "tenant-ghi",
				NodeRole:       "agent",
				NodeStatus:     "degraded",
				NodeVersion:    "v1.1.0",
				NodeGeneration: 3,
				AgentID:        "agent-003",
				NetworkUnitID:  102,
				NetworkAreaID:  202,
				BizID:          302,
				InnerIP:        "192.168.1.12",
				Addressing:     "static",
			},
			NodeConf: &NodeConf{
				PreSetting: map[string]any{
					"maxConnections":   2000,
					"keepAlive":        true,
					"compressionLevel": 5,
				},
				CustomSetting: map[string]any{
					"backupEnabled":  true,
					"backupInterval": "6h",
					"storageQuota":   "10GB",
				},
			},
		},
	}

	return testDatas
}

// TestCreate tests the Create method of the handler.
func (testSuit *TestSuite) TestCreate() {
	type args struct {
		ctx            context.Context
		nodeDeployment *types.NodeDeployment
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
				nodeDeployment: &types.NodeDeployment{
					Token: "666",
					Info: &types.DeploymentInfo{
						OperInstID:         "123",
						BlockingActionName: "wait agent install",
						HostID:             315,
						TenantID:           "single",
						NodeRole:           "agent",
						NodeStatus:         "running",
						NodeVersion:        "v2.1.6-beta.55",
						NodeGeneration:     2,
						AgentID:            "02000000005254001bbe721742528553406b",
						NetworkUnitID:      0,
					},
					NodeConf: &types.NodeConf{
						PreSetting: map[string]any{
							"__BK_NODE_INSTALLER_DISK_REQUIRES__": []map[string]any{
								{
									"demand_mb": 300,
									"dir_path":  "/usr/local",
								},
							},
							"__BK_NODE_INSTALLER_PORT_POLICIES__": []map[string]any{
								{
									"network": "tcp",
									"port":    8009,
								},
							},
							"__BK_NODE_INSTALLER_NETWORK_POLICIES__": []map[string]any{
								{
									"host":    "10.0.0.1",
									"network": "tcp",
									"port":    8006,
								},
							},
						},
						CustomSetting: map[string]any{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			err := testSuit.Handler.Create(tt.args.ctx, tt.args.nodeDeployment)
			if !tt.wantErr {
				testSuit.NoErrorf(err, "Create() error = %v", err)
			}

			testSuit.T().Logf("err: %v", err)
		})
	}
}

// TestHandler_GetInfo ...
func (testSuit *TestSuite) TestHandler_GetInfo() {
	type args struct {
		ctx   context.Context
		Token string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.DeploymentInfo
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   context.Background(),
				Token: "token123456",
			},
			want: &types.DeploymentInfo{
				OperInstID:         "inst-001",
				BlockingActionName: "wait_agent_install",
				HostID:             52296,
				OSType:             "linux",
				TenantID:           "single",
				NodeRole:           "agent",
				NodeStatus:         "running",
				NodeVersion:        "v2.1.6-beta.55",
				NodeGeneration:     2,
				AgentID:            "02000000005254001bbe721742528553406b",
				NetworkUnitID:      0,
				NetworkAreaID:      0,
				BizID:              12,
				InnerIP:            "192.168.1.10",
				Addressing:         "static",
			},
			wantErr: false,
		}}
	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			got, err := testSuit.Handler.GetInfo(tt.args.ctx, tt.args.Token)
			if !tt.wantErr {
				testSuit.Require().NoError(err, "GetInfo() error = %v", err)
			} else {
				testSuit.Require().Error(err, "GetInfo() error = %v", err)
			}

			testSuit.Assert().Equal(got, tt.want, "GetInfo() got = %v, want %v", got, tt.want)
			testSuit.T().Logf("got: %+v", got)
		})
	}
}

// TestHandler_GetNodeConf ...
func (testSuit *TestSuite) TestHandler_GetNodeConf() {
	type args struct {
		ctx   context.Context
		Token string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.NodeConf
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   context.Background(),
				Token: "token123456",
			},
			want: &types.NodeConf{
				PreSetting: map[string]any{
					"maxConnections": int64(1000),
					"timeout":        int64(30),
					"retries":        int64(3),
				},
				CustomSetting: map[string]any{
					"city_id":  "info",
					"run_mode": "agent",
					"cloud_id": int64(0),
					"zone_id":  "1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			got, err := testSuit.Handler.GetNodeConf(tt.args.ctx, tt.args.Token)
			if tt.wantErr {
				testSuit.Require().Error(err, "GetNodeConf() error = %v", err)
			} else {
				testSuit.Require().NoError(err, "GetNodeConf() error = %v", err)
			}

			testSuit.Assert().Equal(got, tt.want, "GetNodeConf() got = %v, want %v", got, tt.want)
			testSuit.T().Logf("got: %+v", got)
		})
	}
}

// TestHandler_SetNodeConf ...
func (testSuit *TestSuite) TestHandler_SetNodeConf() {
	type args struct {
		ctx      context.Context
		Token    string
		nodeConf *types.NodeConf
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   context.Background(),
				Token: "token345678",
				nodeConf: &types.NodeConf{
					PreSetting: map[string]any{
						"__BK_GSE_DATA_AGENT_TLS_CA_FILE__": "ca.crt",
					},
					CustomSetting: map[string]any{
						"run_mode": "agent",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			err := testSuit.Handler.SetNodeConf(tt.args.ctx, tt.args.Token, tt.args.nodeConf)
			if !tt.wantErr {
				testSuit.NoError(err, "SetNodeConf() error = %v", err)
			} else {
				testSuit.Error(err, "SetNodeConf() error = %v", err)
			}
			testSuit.T().Logf("err: %v", err)

			// Verify that the nodeConf was updated correctly
			got, err := testSuit.Handler.GetNodeConf(tt.args.ctx, tt.args.Token)
			testSuit.NoError(err, "GetNodeConf() error = %v", err)
			testSuit.Assert().Equal(got, tt.args.nodeConf, "GetNodeConf() got = %v, want %v", got, tt.args.nodeConf)

			testSuit.T().Logf("got: %+v", got)
		})
	}
}

// TestHandler_UpdateInfo ...
func (testSuit *TestSuite) TestHandler_UpdateInfo() {
	type args struct {
		ctx   context.Context
		token string
		info  *types.DeploymentInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				ctx:   context.Background(),
				token: "token345678",
				info: &types.DeploymentInfo{
					OperInstID:         "inst-003",
					BlockingActionName: "restart",
					HostID:             24680,
					OSType:             "macos",
					TenantID:           "tenant-ghi",
					NodeRole:           "agent",
					NodeStatus:         "degraded",
					NodeVersion:        "v1.1.0",
					NodeGeneration:     2,
					AgentID:            "agent-003",
					NetworkUnitID:      102,
					NetworkAreaID:      202,
					BizID:              302,
					InnerIP:            "192.168.1.12",
					Addressing:         "static",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		testSuit.Run(tt.name, func() {
			err := testSuit.Handler.UpdateInfo(tt.args.ctx, tt.args.token, tt.args.info)
			if !tt.wantErr {
				testSuit.NoError(err, "UpdateInfo() error = %v", err)
			} else {
				testSuit.Error(err, "UpdateInfo() error = %v", err)
			}

			testSuit.T().Logf("err: %v", err)

			// Verify that the info was updated correctly
			got, err := testSuit.Handler.GetInfo(tt.args.ctx, tt.args.token)
			testSuit.NoError(err, "GetInfo() error = %v", err)
			testSuit.Assert().Equal(got, tt.args.info, "GetInfo() got = %v, want %v", got, tt.args.info)
			testSuit.T().Logf("got: %+v", got)
		})
	}
}
