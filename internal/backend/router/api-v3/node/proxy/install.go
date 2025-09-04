/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package proxy

import (
	"encoding/base64"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DefaultNodeGeneration default node generation.
const DefaultNodeGeneration = 2

// Install install proxy.
func (h *handler) Install(ctx *restserver.Context) (interface{}, error) {
	req := new(protoBackend.NodeProxyInstallReq)
	if err := ctx.BindJSON(req); err != nil {
		h.logger.ErrorCtxf(ctx, "failed to install proxy, failed to decode request body. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	nodeDeployments, bizIDs, err := h.generateInstallNodeDeployments(ctx, req)
	if err != nil {
		h.logger.Errorf("failed to install proxy, failed to generate node deployments. err: %v", err)
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.manager.LaunchInstallNode(ctx, manager.InstallNodeParam{
		Type:            types.NodeWorkflowTypeInstallProxy,
		BizIDs:          bizIDs,
		Operator:        ctx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		h.logger.ErrorCtxf(ctx, "failed to install proxy: %v", err)
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyInstallResp)
	resp.ConvertWorkflowID(workflowID)

	h.logger.InfoCtxf(ctx, "launched install proxy workflow: %s", workflowID)

	return resp.GetData(), nil
}

// nolint: funlen
func (h *handler) generateInstallNodeDeployments(
	ctx contextx.ITenantContext, req *protoBackend.NodeProxyInstallReq) ([]*types.NodeDeployment, []int64, error) {

	targetVersions := make([]types.TargetVersion, len(req.GetTargetVersion()))
	for idx, version := range req.GetTargetVersion() {
		targetVersions[idx] = types.TargetVersion{
			OsType:  criteria.OSType(version.GetOsType()),
			CPUArch: criteria.CPUArch(version.GetCpuArch()),
			Version: version.GetVersion(),
		}
	}

	// build biz-id.
	bizIDMap := make(map[int64]struct{})
	for _, host := range req.GetHost() {
		bizIDMap[host.GetBkBizId()] = struct{}{}
	}
	bizIDs := conv.MapKeyToSlice(bizIDMap)

	// fetch networkunit.
	networkUnitMap, err := h.fetchNetworkunits(ctx, req.GetHost())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch networkunits: %w", err)
	}

	// fetch host.
	existedHostMap, err := h.fetchExistedHosts(ctx, req.GetHost())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch existed hosts: %w", err)
	}

	gp := gopool.NewPool()
	nodeDeployments := make([]*types.NodeDeployment, len(req.Host))
	for i := range req.GetHost() {
		idx := i
		reqHost := req.GetHost()[idx]

		gp.Go(func() error {
			networkUnit, ok := networkUnitMap[reqHost.GetBkNetworkunitId()]
			if !ok {
				return fmt.Errorf("failed to find networkunit with id: %d", reqHost.GetBkNetworkunitId())
			}

			loginCreditID := ""
			if existedHost, ok := existedHostMap[reqHost.GetBkHostId()]; ok {
				loginCreditID = existedHost.Dynamic.LoginCreditID
			}

			nodeDeployment := types.NewNodeDeployment(
				&types.DeploymentInfo{
					Host: types.Host{
						HostID:   reqHost.GetBkHostId(),
						TenantID: ctx.TenantID(),
						Static: &types.HostStatic{
							BizID:         reqHost.GetBkBizId(),
							NetworkAreaID: networkUnit.NetworkAreaID,
							InnerIP:       reqHost.GetBkHostInnerip(),
							InnerIPV6:     reqHost.GetBkHostInneripV6(),
							OSType:        reqHost.GetOsType(),
							Addressing:    types.Addressing(reqHost.GetBkAddressing()),
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleProxy,
							NodeStatus:     types.NodeStatusInit,
							NodeGeneration: DefaultNodeGeneration,
							NetworkUnitID:  networkUnit.ID,
							ProxyTags:      types.StringListToProxyTagList(reqHost.GetProxyTags()),
							LoginIP:        reqHost.GetLoginIp(),
							LoginPort:      reqHost.GetLoginPort(),
							LoginUser:      reqHost.GetLoginUser(),
							LoginMode:      types.LoginMode(reqHost.GetLoginMode()),
							LoginCreditID:  loginCreditID,
							ExportIP:       reqHost.GetExportIp(),
							AdvertiseIP:    reqHost.GetAdvertiseIp(),
						},
					},
					CurrentVersionSupports: types.DeploymentVersionSupports{},
					InstallOptions: types.DeploymentInstallOptions{
						ReRegister: reqHost.GetReRegister(),
					},
					UpgradeOptions:  types.DeploymentUpgradeOptions{},
					RestartOptions:  types.DeploymentRestartOptions{},
					TransferOptions: types.DeploymentTransferOptions{},
					TargetVersion:   targetVersions,
				})

			if err = h.processHostCredit(ctx, &nodeDeployment.Info.Host, reqHost.GetLoginPassword(), reqHost.GetLoginKeyFile()); err != nil {
				return fmt.Errorf("failed to process host credit: %w", err)
			}

			nodeDeployments[idx] = nodeDeployment

			return nil
		})
	}
	if err := gp.Wait(); err != nil {
		return nil, nil, err
	}

	return nodeDeployments, bizIDs, nil
}

func (h *handler) fetchNetworkunits(ctx contextx.IContext, hosts []*protoBackend.NodeProxyInstallHost) (map[int64]*types.NetworkUnit, error) {
	networkUnitIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		networkUnitIDMap[host.GetBkNetworkunitId()] = struct{}{}
	}

	networkUnitList, _, err := h.storageNetworkUnit.ListNetworkUnit(ctx, types.UnlimitedPage(), &types.NetworkUnitCondition{
		ExactInclude: &types.NetworkUnitExactFields{
			NetworkUnitID: conv.MapKeyToSlice(networkUnitIDMap),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch networkunit: %w", err)
	}

	networkUnitMap := make(map[int64]*types.NetworkUnit)
	for _, networkUnit := range networkUnitList {
		networkUnitMap[networkUnit.ID] = networkUnit
	}

	return networkUnitMap, nil
}

func (h *handler) fetchExistedHosts(ctx contextx.ITenantContext, hosts []*protoBackend.NodeProxyInstallHost) (map[int64]*types.Host, error) {
	hostIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		if hostID := host.GetBkHostId(); hostID >= 0 {
			hostIDMap[hostID] = struct{}{}
		}
	}

	existedHostList, _, err := h.storageHost.ListHost(ctx, types.UnlimitedPage(), &types.HostCondition{
		ExactInclude: &types.HostExactFields{
			HostID: conv.MapKeyToSlice(hostIDMap),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existed host: %w", err)
	}
	existedHostMap := make(map[int64]*types.Host)
	for _, host := range existedHostList {
		existedHostMap[host.HostID] = host
	}

	return existedHostMap, nil
}

func (h *handler) processHostCredit(ctx contextx.ITenantContext, host *types.Host, password, keyfile string) error {
	var err error
	switch host.Dynamic.LoginMode {
	case types.LoginModeKeyFile:
		if keyfile == "" {
			if host.Dynamic.LoginCreditID == "" {
				err := fmt.Errorf("keyfile is empty and there is not login credit to use. host-id(%d), inner-ip(%s)", host.HostID, host.Static.InnerIP)
				h.logger.ErrorCtxf(ctx, "failed to process host credit: %v", err)

				return err
			}

			// use old credit id.
			return nil
		}

		loginKeyFile, err := base64.StdEncoding.DecodeString(keyfile)
		if err != nil {
			h.logger.Errorf("use base64 decode key file failed, err: %v", err)

			return fmt.Errorf("failed to decode key file, err: %w", err)
		}

		host.Dynamic.LoginCreditID, err = h.storageHostCredit.CreateHostCredit(
			ctx,
			loginKeyFile,
		)
		if err != nil {
			return fmt.Errorf("failed to gen node deployment: %w", err)
		}
		return nil

	case types.LoginModePassword:
		if password == "" {
			if host.Dynamic.LoginCreditID == "" {
				err := fmt.Errorf("password is empty and there is not login credit to use. host-id(%d), inner-ip(%s)", host.HostID, host.Static.InnerIP)
				h.logger.ErrorCtxf(ctx, "failed to process host credit: %v", err)

				return err
			}

			// use old credit id.
			return nil
		}

		host.Dynamic.LoginCreditID, err = h.storageHostCredit.CreateHostCredit(
			ctx,
			[]byte(password),
		)
		if err != nil {
			return fmt.Errorf("failed to gen node deployment: %w", err)
		}
		return nil

	case types.LoginModePasswordVault:
		// notice: password vault don't need to store password.
		return nil

	default:
		err = fmt.Errorf("unsupported this login mode. login-mode(%s)", host.Dynamic.LoginMode)
		h.logger.Error(err)

		return err
	}
}
