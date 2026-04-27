// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { NodeProxyInstallReq, NodeProxyInstallResp, NodeProxyUpgradeReq, NodeProxyUpgradeResp, NodeProxyRestartReq, NodeProxyRestartResp, NodeProxyReconfigReq, NodeProxyReconfigResp, NodeProxyUpdateReq, NodeProxyUpdateResp, NodeProxyUninstallReq, NodeProxyUninstallResp, NodeProxyInstallCheckReq, NodeProxyInstallCheckResp, NodeProxyUpgradeCheckReq, NodeProxyUpgradeCheckResp, NodeProxyAssignUnitReq, NodeProxyAssignUnitResp } from '@/@types/node_proxy';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const NodeProxyService = {
  // NodeProxyInstall installs node proxy.
  NodeProxyInstall: async <Request = NodeProxyInstallReq, ResponseData = NodeProxyInstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/install')(params, config),
  // NodeProxyUpgrade upgrades node proxy.
  NodeProxyUpgrade: async <Request = NodeProxyUpgradeReq, ResponseData = NodeProxyUpgradeResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/upgrade')(params, config),
  // NodeProxyRestart restarts node proxy.
  NodeProxyRestart: async <Request = NodeProxyRestartReq, ResponseData = NodeProxyRestartResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/restart')(params, config),
  // NodeProxyReconfig reconfigs node proxy.
  NodeProxyReconfig: async <Request = NodeProxyReconfigReq, ResponseData = NodeProxyReconfigResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/reconfig')(params, config),
  // NodeProxyUpdate updates node proxy.
  NodeProxyUpdate: async <Request = NodeProxyUpdateReq, ResponseData = NodeProxyUpdateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/update')(params, config),
  // NodeProxyUninstall uninstall node proxy.
  NodeProxyUninstall: async <Request = NodeProxyUninstallReq, ResponseData = NodeProxyUninstallResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/uninstall')(params, config),
  // NodeProxyInstallCheck checks node proxy install.
  NodeProxyInstallCheck: async <Request = NodeProxyInstallCheckReq, ResponseData = NodeProxyInstallCheckResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/install_check')(params, config),
  // NodeProxyUpgradeCheck checks node proxy upgrade.
  NodeProxyUpgradeCheck: async <Request = NodeProxyUpgradeCheckReq, ResponseData = NodeProxyUpgradeCheckResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/upgrade_check')(params, config),
  // NodeProxyAssignUnit batch-assigns a network unit to unassigned proxy hosts.
  NodeProxyAssignUnit: async <Request = NodeProxyAssignUnitReq, ResponseData = NodeProxyAssignUnitResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/proxy/assign_unit')(params, config),
};

