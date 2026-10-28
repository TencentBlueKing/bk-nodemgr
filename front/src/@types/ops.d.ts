// gen-api.js 自动生成，请勿手动修改
// NodeOpsHistoryInfo describes one batch of out-of-band settings.
export interface NodeOpsHistoryInfo {
  operate_id: string;
  operator: string;
  operate_time: number;
  node_role: string;
}

// NodeOpsHistoryExactConditions describes the exact conditions of the history
// list request.
// Values within a field are ORed; different fields are ANDed.
// Empty fields do not restrict the query.
export interface NodeOpsHistoryExactConditions {
  operate_id: string[];
  operator: string[];
  node_role: string[];
}

// NodeOpsHistoryFuzzyConditions describes the fuzzy conditions of the history
// list request.
export interface NodeOpsHistoryFuzzyConditions {
}

// NodeOpsHistoryListReq describes the history list request for all node roles.
export interface NodeOpsHistoryListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: NodeOpsHistoryExactConditions;
  fuzzy_include_conditions: NodeOpsHistoryFuzzyConditions;
  operate_time_range: TimeRange;
}

// NodeOpsHistoryListResp describes the operation history list response.
export interface NodeOpsHistoryListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeOpsHistoryListRespData;
}

export interface NodeOpsHistoryListRespData {
  // Number of matching operations, not hosts.
  total: number;
  items: NodeOpsHistoryInfo[];
}

// NodeOpsHistoryHostSnapshot preserves host information and submitted settings.
export interface NodeOpsHistoryHostSnapshot {
  operate_id: string;
  bk_host_id: number;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
  bk_biz_id: number;
  bk_biz_name: string;
  bk_networkarea_id: number;
  bk_networkarea_name: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  ops_console_host_id: number;
  ops_console_host_innerip_list: string[];
  ops_console_host_innerip_v6_list: string[];
  ops_out_band_ip: string;
  ops_out_band_port: number;
  ops_out_band_protocol: string;
  ops_out_band_type: string;
}

// NodeOpsHistoryDetailReq describes the snapshot retrieval request for all node
// roles.
export interface NodeOpsHistoryDetailReq {
  // Required operation ID. Returns all host snapshots without pagination.
  operate_id: string;
}

// NodeOpsHistoryDetailResp describes the read-only snapshot retrieval response.
export interface NodeOpsHistoryDetailResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: NodeOpsHistoryDetailRespData;
}

export interface NodeOpsHistoryDetailRespData {
  // Number of accessible host snapshots for this operation.
  total: number;
  items: NodeOpsHistoryHostSnapshot[];
}

