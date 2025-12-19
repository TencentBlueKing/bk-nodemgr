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
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
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

// creditExpiredInterval is the expiration interval for proxy host credit.
const creditExpiredInterval = time.Hour * 24

// genCreditExpiredAt generates the expiration time for proxy host credit.
func genCreditExpiredAt(expiredIntervalSec int64) time.Time {
	if expiredIntervalSec > 0 {
		return time.Now().Add(time.Duration(expiredIntervalSec) * time.Second)
	}

	return time.Now().Add(creditExpiredInterval)
}

// Install install proxy.
func (h *handler) Install(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeProxyInstallReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install proxy, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	nodeDeployments, bizIDs, err := h.generateInstallNodeDeployments(rCtx, req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install proxy, failed to generate node deployments")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	workflowID, err := h.nodeMgrIface.LaunchInstallNode(rCtx, types.InstallNodeParam{
		Type:            types.NodeWorkflowTypeInstallProxy,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install proxy")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeProxyInstallResp)
	resp.ConvertWorkflowID(workflowID)

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched install proxy workflow")

	return resp.GetData(), nil
}

// nolint: funlen
func (h *handler) generateInstallNodeDeployments(
	nCtx contextx.IContext, req *protoBackend.NodeProxyInstallReq) ([]*types.NodeDeployment, []int64, error) {

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
	networkUnitMap, err := h.fetchNetworkunits(nCtx, req.GetHost())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch networkunits: %w", err)
	}

	// fetch host.
	existedHostMap, err := h.fetchExistedHosts(nCtx, req.GetHost())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch existed hosts: %w", err)
	}

	gp := gopool.NewPool()
	nodeDeployments := make([]*types.NodeDeployment, len(req.GetHost()))
	for i := range req.GetHost() {
		idx := i
		reqHost := req.GetHost()[idx]

		gp.Go(func() error {
			networkUnit, ok := networkUnitMap[reqHost.GetBkNetworkunitId()]
			if !ok {
				return fmt.Errorf("failed to find networkunit with id: %d", reqHost.GetBkNetworkunitId())
			}

			installOriginUnit, ok := networkUnitMap[reqHost.GetProxyInstallOriginUnitId()]
			if !ok {
				return fmt.Errorf("failed to find install origin networkunit with id: %d", reqHost.GetProxyInstallOriginUnitId())
			}

			loginCreditID := ""
			if existedHost, ok := existedHostMap[reqHost.GetBkHostId()]; ok {
				loginCreditID = existedHost.Dynamic.LoginCreditID
			}

			nodeDeployment := types.NewNodeDeployment(
				&types.DeploymentInfo{
					Host: types.Host{
						HostID:   reqHost.GetBkHostId(),
						TenantID: nCtx.TenantID(),
						Static: &types.HostStatic{
							BizID:         reqHost.GetBkBizId(),
							NetworkAreaID: networkUnit.NetworkAreaID,
							InnerIPList:   []string{reqHost.GetBkHostInnerip()},
							InnerIPV6List: []string{reqHost.GetBkHostInneripV6()},
							OSType:        reqHost.GetOsType(),
							Addressing:    types.Addressing(reqHost.GetBkAddressing()),
						},
						Dynamic: &types.HostDynamic{
							NodeRole:                 types.NodeRoleProxy,
							NodeStatus:               types.NodeStatusInit,
							NodeGeneration:           DefaultNodeGeneration,
							NetworkUnitID:            networkUnit.ID,
							ProxyTags:                types.StringListToProxyTagList(reqHost.GetProxyTags()),
							LoginIP:                  reqHost.GetLoginIp(),
							LoginPort:                reqHost.GetLoginPort(),
							LoginUser:                reqHost.GetLoginUser(),
							LoginMode:                types.LoginMode(reqHost.GetLoginMode()),
							LoginCreditID:            loginCreditID,
							ExportIP:                 reqHost.GetExportIp(),
							ExportIPV6:               reqHost.GetExportIpV6(),
							AdvertiseIP:              reqHost.GetAdvertiseIp(),
							AdvertiseIPV6:            reqHost.GetAdvertiseIpV6(),
							ProxyInstallOriginUnitID: reqHost.GetProxyInstallOriginUnitId(),
						},
					},
					CurrentVersionSupports: types.DeploymentVersionSupports{},
					InstallOptions: types.DeploymentInstallOptions{
						ReRegister:    reqHost.GetReRegister(),
						DirectInstall: installOriginUnit.IsDirect,
						IsManual:      req.GetIsManual(),
					},
					UpgradeOptions:  types.DeploymentUpgradeOptions{},
					RestartOptions:  types.DeploymentRestartOptions{},
					TransferOptions: types.DeploymentTransferOptions{},
					TargetVersion:   targetVersions,
				})

			err = h.processHostCredit(nCtx, &nodeDeployment.Info.Host,
				reqHost.GetLoginPassword(),
				reqHost.GetLoginKeyFile(),
				reqHost.GetCreditExpiredIntervalSec())
			if err != nil {
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

	// add proxy_install_origin_unit_id.
	for _, host := range hosts {
		networkUnitIDMap[host.GetProxyInstallOriginUnitId()] = struct{}{}
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

func (h *handler) fetchExistedHosts(ctx contextx.IContext, hosts []*protoBackend.NodeProxyInstallHost) (map[int64]*types.Host, error) {
	hostIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		if hostID := host.GetBkHostId(); hostID >= 0 {
			hostIDMap[hostID] = struct{}{}
		}
	}

	existedHostList, _, err := h.storageHost.ListHost(ctx, types.UnlimitedPage(), &types.HostCondition{
		StaticExactInclude: &types.HostStaticExactFields{
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

func (h *handler) processHostCredit(nCtx contextx.IContext, host *types.Host, password, keyfile string, creditExpiredIntervalSec int64) error {
	var err error
	switch host.Dynamic.LoginMode {
	case types.LoginModeKeyFile:
		if keyfile == "" {
			if host.Dynamic.LoginCreditID == "" {
				err := fmt.Errorf("keyfile is empty and there is not login credit to use. host-id(%d), inner-ip(%v)", host.HostID, host.Static.InnerIPList)
				logger.G.Biz(nCtx).WithErr(err).With("host-id", host.HostID, "inner-ip", host.Static.InnerIPList).Error("failed to process host credit")

				return err
			}

			// use old credit id.
			return nil
		}

		loginKeyFile, err := base64.StdEncoding.DecodeString(keyfile)
		if err != nil {
			logger.G.Biz(nCtx).WithErr(err).Error("use base64 decode key file failed")

			return fmt.Errorf("failed to decode key file: %w", err)
		}

		host.Dynamic.LoginCreditID, err = h.storageHostCredit.CreateHostCredit(
			nCtx,
			loginKeyFile,
			genCreditExpiredAt(creditExpiredIntervalSec),
		)
		if err != nil {
			return fmt.Errorf("failed to gen node deployment: %w", err)
		}

		return nil

	case types.LoginModePassword:
		if password == "" {
			if host.Dynamic.LoginCreditID == "" {
				err := fmt.Errorf("password is empty and there is not login credit to use. host-id(%d), inner-ip(%v)", host.HostID, host.Static.InnerIPList)
				logger.G.Biz(nCtx).WithErr(err).With("host-id", host.HostID, "inner-ip", host.Static.InnerIPList).Error("failed to process host credit")

				return err
			}

			// use old credit id.
			return nil
		}

		host.Dynamic.LoginCreditID, err = h.storageHostCredit.CreateHostCredit(
			nCtx,
			[]byte(password),
			genCreditExpiredAt(creditExpiredIntervalSec),
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
		logger.G.Biz(nCtx).WithErr(err).With("login-mode", host.Dynamic.LoginMode).Error("unsupported login mode")

		return err
	}
}
