// gen-api.js 自动生成，请勿手动修改
// Page describes the generaic conditions when paging query.
export interface Page {
  offset: number;
  limit: number;
}

// Business describes the business informations.
export interface Business {
  tenantId: string;
  bkBizId: number;
  bkBizName: string;
}

// NetworkArea describes the network area informations.
export interface NetworkArea {
  tenantId: string;
  bkNetworkareaId: number;
  bkNetworkareaName: string;
  bkCloudVendor: number;
}

// Link describes the network unit link points target.
export interface Link {
  tenantId: string;
  apId: number;
}

// AccessPoint describes the access point informations.
export interface AccessPoint {
  tenantId: string;
  apId: number;
  apName: string;
  endpoints: Endpoints;
}

// NetworkUnit describes the network unit informations.
export interface NetworkUnit {
  bkNetworkunitId: number;
  bkNetworkunitName: string;
  accessPoints: AccessPoint[];
  links: Links;
}

// HostState describes the host state informations. Usually contains
// agent-related things.
export interface HostState {
  status: string;
  version: string;
  bkAgentId: string;
}

// HostInfo describes the host info informations. Usually contains static
// configs.
export interface HostInfo {
  bkHostId: number;
  bkHostName: string;
  bkBizId: number;
  bkBizName: string;
  bkSetId: number;
  bkSetName: string;
  bkModuleId: number;
  bkModuleName: string;
  bkNetworkareaId: number;
  bkNetworkareaName: string;
  bkNetworkunitId: number;
  bkNetworkunitName: string;
  bkOperateDeptId: number;
  bkOperateDeptName: string;
  bkHostInnerip: string;
  bkHostInneripV6: string;
  bkHostOuterip: string;
  bkHostOuteripV6: string;
  loginIp: string;
  loginPort: number;
  bkOsType: number;
}

// Host describes the host informations.
export interface Host {
  state: HostState;
  info: HostInfo;
  createAt: number;
  updatedAt: number;
}

