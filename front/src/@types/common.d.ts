// gen-api.js 自动生成，请勿手动修改
// Page describes the generaic conditions when paging query.
interface Page {
  offset: number;
  limit: number;
}

// Business describes the business informations.
interface Business {
  tenant_id: string;
  bk_biz_id: number;
  bk_biz_name: string;
}

// NetworkArea describes the network area informations.
interface NetworkArea {
  tenant_id: string;
  bk_networkarea_id: number;
  bk_networkarea_name: string;
  bk_cloud_vendor: string;
}

// Link describes the network unit link points target.
interface Link {
  accesspoint_id: number;
}

// Links describes the network unit links.
interface Links {
  cluster: Link;
  file: Link;
  data: Link;
}

// AccessPoint describes the access point informations.
interface AccessPoint {
  tenant_id: string;
  accesspoint_id: number;
  accesspoint_name: string;
  endpoints: AccessPointEndpoints;
}

interface AccessPointEndpoints {
  cluster: string[];
  file: string[];
  data: string[];
}

// NetworkUnit describes the network unit informations.
interface NetworkUnit {
  tenant_id: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  access_points: AccessPoint[];
  links: Links;
}

// NetworkUnitBrief describes the network unit brief informations.
// only contains the access point ids.
interface NetworkUnitBrief {
  tenant_id: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  access_points: number[];
  links: Links;
}

// NetworkUnitBrief describes the network unit brief informations.
// only contains the access point ids.
export interface NetworkUnitBrief {
  tenantId: string;
  bkNetworkunitId: number;
  bkNetworkunitName: string;
  bkNetworkareaId: number;
  accessPoints: number[];
  links: Links;
}

// HostState describes the host state informations. Usually contains
// agent-related things.
interface HostState {
  node_role: string;
  node_status: string;
  node_version: string;
  bk_agent_id: string;
}

// HostInfo describes the host info informations. Usually contains static
// configs.
interface HostInfo {
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_host_name: string;
  dept_name: string;
  bk_host_innerip: string;
  bk_host_innerip_v6: string;
  bk_host_outerip: string;
  bk_host_outerip_v6: string;
  bk_mac: string;
  bk_os_type: string;
}

// Host describes the host informations.
interface Host {
  tenant_id: string;
  bk_host_id: number;
  info: HostInfo;
  state: HostState;
  create_at: number;
  updated_at: number;
}

