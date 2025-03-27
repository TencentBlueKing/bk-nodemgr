/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package topo

import (
	"context"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/tenant"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

var gsePrepareOnce = sync.Once{}
var gsePrepareDirectUnitID int64
var gsePrepareNormalUnitID int64

func prepareGSEData(t *testing.T, ctx context.Context) {
	gsePrepareOnce.Do(func() {
		tenantID, _ := tenant.GetID(ctx)

		// pre insert.
		h := testClient(t)

		var apResult *AccessPointResult
		var err error
		gsePrepareDirectUnitID, apResult, err = h.CreateNetworkUnit(ctx, &types.NetworkUnit{
			TenantID: tenantID,
			Name:     "direct-unit",
			IsDirect: true,
			DirectEndpoints: &types.Endpoints{
				Cluster: []string{"1.1.1.1:20001", "2.2.2.2:20001"},
				File:    []string{"3.3.3.3:20002"},
				Data:    []string{"4.4.4.4:20003"},
			},
		}, &types.AccessPoint{
			TenantID:      tenantID,
			NetworkAreaID: 0,
			Name:          "direct-access",
			Endpoints: types.Endpoints{
				Cluster: []string{"5.5.5.5", "6.6.6.6"},
				File:    []string{"7.7.7.7"},
				Data:    []string{"8.8.8.8"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		directAccessPointID := apResult.Created[0].ID
		gsePrepareNormalUnitID, _, err = h.CreateNetworkUnit(ctx, &types.NetworkUnit{
			TenantID: tenantID,
			Name:     "normal-unit",
			IsDirect: false,
			Links: types.Links{
				Cluster: &types.Link{NetworkAreaID: 0, NetworkUnitID: gsePrepareDirectUnitID, AccessPointID: directAccessPointID},
				File:    &types.Link{NetworkAreaID: 0, NetworkUnitID: gsePrepareDirectUnitID, AccessPointID: directAccessPointID},
				Data:    &types.Link{NetworkAreaID: 0, NetworkUnitID: gsePrepareDirectUnitID, AccessPointID: directAccessPointID},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		err = h.UpsertManyHost(ctx, &types.Host{
			HostID:   10001,
			TenantID: tenantID,
			Static: &types.HostStatic{
				BizID:         0,
				NetworkAreaID: 0,
				InnerIP:       "9.9.9.9,10.10.10.10",
			},
			Dynamic: &types.HostDynamic{
				NetworkUnitID:    gsePrepareNormalUnitID,
				NodeRole:         types.NodeRoleProxy,
				ProxyClusterPort: 12001,
				ProxyFilePort:    12002,
				ProxyDataPort:    12003,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

// Test_storage_GetAgentV4AccessEndpoints test get agent access endpoints.
func Test_storage_GetAgentV4AccessEndpoints(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareGSEData(t, ctx)

	type args struct {
		ctx    context.Context
		unitID int64
	}
	tests := []struct {
		name          string
		args          args
		wantClusterEp []string
		wantFileEp    []string
		wantDataEp    []string
		wantErr       bool
	}{
		{
			name: "direct",
			args: args{
				ctx:    ctx,
				unitID: gsePrepareDirectUnitID,
			},
			wantClusterEp: []string{"1.1.1.1:20001", "2.2.2.2:20001"},
			wantFileEp:    []string{"3.3.3.3:20002"},
			wantDataEp:    []string{"4.4.4.4:20003"},
			wantErr:       false,
		},
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				unitID: gsePrepareNormalUnitID,
			},
			wantClusterEp: []string{"9.9.9.9:12001", "10.10.10.10:12001"},
			wantFileEp:    []string{"9.9.9.9:12002", "10.10.10.10:12002"},
			wantDataEp:    []string{"9.9.9.9:12003", "10.10.10.10:12003"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			clusterEp, fileEp, dataEp, err := s.GetV4AgentAccessEndpoints(tt.args.ctx, tt.args.unitID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAgentV4AccessEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, ep := range tt.wantClusterEp {
				if !slices.Contains(clusterEp, ep) {
					t.Errorf("GetAgentV4AccessEndpoints() clusterEp = %v, want %v", clusterEp, tt.wantClusterEp)
				}
			}

			for _, ep := range tt.wantFileEp {
				if !slices.Contains(fileEp, ep) {
					t.Errorf("GetAgentV4AccessEndpoints() fileEp = %v, want %v", fileEp, tt.wantFileEp)
				}
			}

			for _, ep := range tt.wantDataEp {
				if !slices.Contains(dataEp, ep) {
					t.Errorf("GetAgentV4AccessEndpoints() dataEp = %v, want %v", dataEp, tt.wantDataEp)
				}
			}
		})
	}
}

// Test_storage_GetProxyUpstreamAccessEndpoints test get proxy upstream endpoints.
func Test_storage_GetProxyUpstreamAccessEndpoints(t *testing.T) {
	tenant.SetMode(tenant.ModeMultiple)
	ctx, _ := tenant.SetID(context.Background(), "test")

	prepareGSEData(t, ctx)

	type args struct {
		ctx    context.Context
		unitID int64
	}
	tests := []struct {
		name          string
		args          args
		wantClusterEp []string
		wantFileEp    []string
		wantDataEp    []string
		wantErr       bool
	}{
		{
			name: "direct",
			args: args{
				ctx:    ctx,
				unitID: gsePrepareDirectUnitID,
			},
			wantErr: true,
		},
		{
			name: "normal",
			args: args{
				ctx:    ctx,
				unitID: gsePrepareNormalUnitID,
			},
			wantClusterEp: []string{"5.5.5.5", "6.6.6.6"},
			wantFileEp:    []string{"7.7.7.7"},
			wantDataEp:    []string{"8.8.8.8"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testClient(t)
			clusterEp, fileEp, dataEp, err := s.GetProxyUpstreamAccessEndpoints(tt.args.ctx, tt.args.unitID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAgentAccessEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(clusterEp, tt.wantClusterEp) {
				t.Errorf("GetAgentAccessEndpoints() clusterEp = %v, want %v", clusterEp, tt.wantClusterEp)
			}

			if !reflect.DeepEqual(fileEp, tt.wantFileEp) {
				t.Errorf("GetAgentAccessEndpoints() fileEp = %v, want %v", fileEp, tt.wantFileEp)
			}

			if !reflect.DeepEqual(dataEp, tt.wantDataEp) {
				t.Errorf("GetAgentAccessEndpoints() dataEp = %v, want %v", dataEp, tt.wantDataEp)
			}
		})
	}
}
