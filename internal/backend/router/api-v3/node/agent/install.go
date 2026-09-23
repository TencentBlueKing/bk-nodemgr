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

// Package agent ...
package agent

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/manager/compatibility"
	authRouter "github.com/TencentBlueKing/bk-nodemgr/internal/backend/router/api-v3/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/crypter"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/gopool"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// DefaultNodeGeneration default node generation.
const DefaultNodeGeneration = 2

// agentHostCreditExpiredInterval is the expiration interval for agent host credit.
const agentHostCreditExpiredInterval = time.Hour * 24

// generateHostCreditExpiredAt generates the expiration time for agent host credit.
func generateHostCreditExpiredAt() time.Time {
	return time.Now().Add(agentHostCreditExpiredInterval)
}

// AgentInstall install agent.
func (h *handler) AgentInstall(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.NodeAgentInstallReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install agent, failed to decode request body")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	nodeDeployments, bizIDs, err := h.generateInstallNodeDeployments(rCtx, req)
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install agent, failed to generate node deployments. err")
		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	networkUnitIDMap := make(map[int64]struct{})
	for _, host := range req.GetHost() {
		networkUnitIDMap[host.GetBkNetworkunitId()] = struct{}{}
	}
	networkUnitIDs := conv.MapKeyToSlice(networkUnitIDMap)
	networkUnitResources := authRouter.BuildNetworkUnitResources(networkUnitIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionNetworkUnitUseForAgent, networkUnitResources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to install agent, network unit permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	resources := authRouter.BuildBizResources(bizIDs...)
	if authErr := h.authorizer.Check(rCtx, auth.ActionAgentOperate, resources); authErr != nil {
		logger.G.Biz(rCtx).WithErr(authErr).Error("failed to install agent, permission denied")
		return nil, resterrf.ErrWrap(resterrf.PermissionDenied, authErr)
	}

	workflowID, err := h.nodeMgrIface.LaunchInstallNode(rCtx, types.InstallNodeParam{
		Type:            types.NodeWorkflowTypeInstallAgent,
		BizIDs:          bizIDs,
		Operator:        rCtx.BKUsername(),
		NodeDeployments: nodeDeployments,
	})
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to install agent")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.NodeAgentInstallResp)
	resp.ConvertWorkflowID(workflowID)

	logger.G.Biz(rCtx).With("workflow-id", workflowID).Info("launched install agent workflow")

	return resp.GetData(), nil
}

func (h *handler) readCompatibilityModePolicy(nCtx contextx.IContext) compatibility.Policy {
	if h.storageGlobalSettings == nil {
		logger.G.Biz(nCtx).Warn("global settings storage is unavailable, use default plugin compatibility mode policy")
		return compatibility.ParseStoredPolicy("", nCtx)
	}

	raw, err := h.storageGlobalSettings.GetGlobalSetting(nCtx, globalsettings.PluginCompatibilityModePolicy)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Warn("failed to read plugin compatibility mode policy, use default policy")
		return compatibility.ParseStoredPolicy("", nCtx)
	}

	return compatibility.ParseStoredPolicy(raw, nCtx)
}

// nolint: funlen, gocognit
func (h *handler) generateInstallNodeDeployments(
	nCtx contextx.IContext, req *protoBackend.NodeAgentInstallReq) ([]*types.NodeDeployment, []int64, error) {

	// if this install is manual
	isManual := req.GetIsManual()
	reqHosts := req.GetHost()

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
	for _, host := range reqHosts {
		bizIDMap[host.GetBkBizId()] = struct{}{}
	}
	bizIDs := conv.MapKeyToSlice(bizIDMap)

	// fetch networkunit.
	networkUnitMap, err := h.fetchNetworkunits(nCtx, reqHosts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch networkunits: %w", err)
	}

	compatibilityPolicy := h.readCompatibilityModePolicy(nCtx)

	// fetch host.
	existedHostMap := make(map[int64]*types.Host)
	if !isManual {
		existedHostMap, err = h.fetchExistedHosts(nCtx, reqHosts)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch existed hosts: %w", err)
		}
	}

	var credentialCrypter crypter.Crypter
	if !isManual {
		credentialCrypter, err = h.initCredentialCrypter(nCtx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to init credential crypter for password hosts: %w", err)
		}
	}

	gp := gopool.NewPool()
	nodeDeployments := make([]*types.NodeDeployment, len(reqHosts))
	for i := range reqHosts {
		idx := i
		reqHost := reqHosts[idx]

		gp.Go(func() error {
			networkUnit, ok := networkUnitMap[reqHost.GetBkNetworkunitId()]
			if !ok {
				return fmt.Errorf("failed to find networkunit with id: %d", reqHost.GetBkNetworkunitId())
			}

			installMethod := types.NodeInstallMethod(reqHost.GetInstallMethod())
			osType := criteria.OSType(reqHost.GetOsType())
			if err := installMethod.CheckAvailable(osType); err != nil {
				return fmt.Errorf("failed to check install_method availability: %w", err)
			}

			loginCreditID := ""
			existedHost, ok := existedHostMap[reqHost.GetBkHostId()]
			if ok {
				loginCreditID = existedHost.Dynamic.LoginCreditID
			}
			logger.G.Biz(nCtx).
				With("host-id", reqHost.GetBkHostId(), "inner-ip", reqHost.GetBkHostInnerip(), "host-exited", ok).Info("generating node deployment")

			nodeDeployment := types.NewNodeDeployment(
				&types.DeploymentInfo{
					Host: types.Host{
						HostID:   reqHost.GetBkHostId(),
						TenantID: nCtx.TenantID(),
						Static: &types.HostStatic{
							BizID:         reqHost.GetBkBizId(),
							NetworkAreaID: networkUnit.NetworkAreaID,
							InnerIPList:   reqHost.GetBkHostInnerip(),
							InnerIPV6List: reqHost.GetBkHostInneripV6(),
							OSType:        osType.String(),
							Addressing:    types.Addressing(reqHost.GetBkAddressing()),
						},
						Dynamic: &types.HostDynamic{
							NodeRole:       types.NodeRoleAgent,
							NodeStatus:     types.NodeStatusInit,
							NodeGeneration: DefaultNodeGeneration,
							NetworkUnitID:  networkUnit.ID,
							LoginIP:        reqHost.GetLoginIp(),
							LoginPort:      reqHost.GetLoginPort(),
							LoginUser:      reqHost.GetLoginUser(),
							LoginMode:      types.LoginMode(reqHost.GetLoginMode()),
							LoginCreditID:  loginCreditID,
						},
					},
					CurrentVersionSupports: types.DeploymentVersionSupports{},
					InstallOptions: types.DeploymentInstallOptions{
						ReRegister:               reqHost.GetReRegister(),
						RenewGSETask:             reqHost.GetRenewGseTask(),
						RenewGSEProc:             reqHost.GetRenewGseProc(),
						InstallPreOrderedPlugins: reqHost.GetInstallPreOrderedPlugins(),
						DirectInstall:            networkUnit.IsDirect,
						InstallMethod:            installMethod,
						EnableCompatibilityMode: compatibility.DecideCompatibilityMode(
							compatibilityPolicy,
							nCtx.TenantID(),
							reqHost.GetBkBizId(),
							"bkmonitorbeat",
						),
						IsManual: isManual,
					},
					UpgradeOptions:  types.DeploymentUpgradeOptions{},
					RestartOptions:  types.DeploymentRestartOptions{},
					TransferOptions: types.DeploymentTransferOptions{},
					TargetVersion:   targetVersions,
				})

			if !isManual {
				if err = h.processHostCredit(
					nCtx, &nodeDeployment.Info.Host, reqHost.GetLoginPassword(), reqHost.GetLoginKeyFile(), credentialCrypter,
				); err != nil {
					return fmt.Errorf("failed to process host credit: %w", err)
				}
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

// initCredentialCrypter builds the credential crypter of the globally
// configured suite: CLASSIC uses the default RSA keypair (ciphertext v1/v2),
// SHANGMI uses the default SM2 keypair (ciphertext v3). Ciphertexts of the
// disabled suite are rejected.
func (h *handler) initCredentialCrypter(nCtx contextx.IContext) (crypter.Crypter, error) {
	keyType := h.cryptoType.CredentialKeyType()

	if err := h.storageCipher.EnsureDefaultCipher(nCtx, keyType); err != nil {
		return nil, fmt.Errorf("failed to ensure default cipher: %w", err)
	}

	cipher, err := h.storageCipher.GetCipher(nCtx, types.DefaultCipherName, keyType)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s cipher: %w", keyType, err)
	}

	if cipher == nil || len(cipher.PrivateKey) == 0 {
		return nil, fmt.Errorf("%s private key is unavailable", keyType)
	}

	switch keyType {
	case types.CipherKeyTypeSM2:
		return crypter.NewSM2CrypterFromPrivateKey(cipher.PrivateKey)
	default:
		return crypter.NewRSACrypterFromPrivateKey(cipher.PrivateKey)
	}
}

func (h *handler) fetchNetworkunits(nCtx contextx.IContext, hosts []*protoBackend.NodeAgentInstallReq_Host) (map[int64]*types.NetworkUnit, error) {
	networkUnitIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		if networkUnitID := host.GetBkNetworkunitId(); networkUnitID >= 0 {
			networkUnitIDMap[networkUnitID] = struct{}{}
		}
	}

	if len(networkUnitIDMap) == 0 {
		return make(map[int64]*types.NetworkUnit), nil
	}

	networkUnitList, _, err := h.storageNetworkUnit.ListNetworkUnit(nCtx, types.UnlimitedPage(), &types.NetworkUnitCondition{
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

func (h *handler) fetchExistedHosts(nCtx contextx.IContext, hosts []*protoBackend.NodeAgentInstallReq_Host) (map[int64]*types.Host, error) {
	hostIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		if hostID := host.GetBkHostId(); hostID >= 0 {
			hostIDMap[hostID] = struct{}{}
		}
	}

	if len(hostIDMap) == 0 {
		return make(map[int64]*types.Host), nil
	}

	existedHostList, _, err := h.storageHost.ListHost(nCtx, types.UnlimitedPage(), &types.HostCondition{
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

func (h *handler) processHostCredit(nCtx contextx.IContext, host *types.Host, password, keyfile string, credentialCrypter crypter.Crypter) error {
	switch host.Dynamic.LoginMode {
	case types.LoginModeKeyFile:
		return h.processKeyFileCredit(nCtx, host, keyfile, credentialCrypter)

	case types.LoginModePassword:
		return h.processPasswordCredit(nCtx, host, password, credentialCrypter)

	case types.LoginModePasswordVault:
		// notice: password vault don't need to store password.
		return nil

	default:
		err := fmt.Errorf("unsupported this login mode. login-mode(%s)", host.Dynamic.LoginMode)
		logger.G.Biz(nCtx).WithErr(err).With("login-mode", host.Dynamic.LoginMode).Error("unsupported login mode")

		return err
	}
}

func (h *handler) processKeyFileCredit(nCtx contextx.IContext, host *types.Host, keyfile string, credentialCrypter crypter.Crypter) error {
	if keyfile == "" {
		if host.Dynamic.LoginCreditID == "" {
			err := fmt.Errorf("keyfile is empty and there is not login credit to use. host-id(%d), inner-ip(%v)", host.HostID, host.Static.InnerIPList)
			logger.G.Biz(nCtx).WithErr(err).Error("failed to process host credit")

			return err
		}

		// use old credit id.
		return nil
	}

	plainKeyFile, err := crypter.DecryptBase64Ciphertext(credentialCrypter, keyfile)
	if err != nil {
		return fmt.Errorf("failed to decrypt key file: %w", err)
	}

	loginKeyFile, err := base64.StdEncoding.DecodeString(plainKeyFile)
	if err != nil {
		logger.G.Biz(nCtx).WithErr(err).Error("use base64 decode key file failed")

		return fmt.Errorf("failed to decode key file: %w", err)
	}

	host.Dynamic.LoginCreditID, err = h.storageHostCredit.CreateHostCredit(nCtx, loginKeyFile, generateHostCreditExpiredAt())
	if err != nil {
		return fmt.Errorf("failed to create host credit: %w", err)
	}

	return nil
}

func (h *handler) processPasswordCredit(nCtx contextx.IContext, host *types.Host, password string, credentialCrypter crypter.Crypter) error {
	if password == "" {
		if host.Dynamic.LoginCreditID == "" {
			err := fmt.Errorf("password is empty and there is not login credit to use. host-id(%d), inner-ip(%v)", host.HostID, host.Static.InnerIPList)
			logger.G.Biz(nCtx).WithErr(err).Error("failed to process host credit")

			return err
		}

		// use old credit id.
		return nil
	}

	// decrypt password.
	plainPassword, err := crypter.DecryptBase64Ciphertext(credentialCrypter, password)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	host.Dynamic.LoginCreditID, err = h.storageHostCredit.CreateHostCredit(
		nCtx, []byte(plainPassword), generateHostCreditExpiredAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to create host credit: %w", err)
	}

	return nil
}
