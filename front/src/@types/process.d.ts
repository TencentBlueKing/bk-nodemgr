// gen-api.js 自动生成，请勿手动修改
// ProcessExactConditions describes the exact match conditions for process
// query.
export interface ProcessExactConditions {
  bk_host_id: number[];
  bk_biz_id: number[];
  plugin_group: string[];
  generation: number[];
  platform_os: string[];
  platform_arch: string[];
  status: string[];
  agent_id: string[];
  version: string[];
  plugin_name: string[];
  plugin_pkg_name: string[];
}

// ProcessFuzzyConditions describes the fuzzy match conditions for process
// query.
export interface ProcessFuzzyConditions {
  name: string[];
  plugin_pkg_name: string[];
}

// ProcessListReq describes the process list request.
export interface ProcessListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: ProcessExactConditions;
  fuzzy_include_conditions: ProcessFuzzyConditions;
}

// ProcessListResp describes the process list response.
export interface ProcessListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: ProcessListRespData;
}

export interface ProcessListRespData {
  total: number;
  items: Process[];
}

// GetProcessDistributionByHostIDReq describes the process statistics request.
export interface GetProcessDistributionByHostIDReq {
  exact_include_conditions: ProcessExactConditions;
  fuzzy_include_conditions: ProcessFuzzyConditions;
}

// GetProcessDistributionByHostIDResp describes the process statistics response.
export interface GetProcessDistributionByHostIDResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: Record<int64, number>;
}

// GetProcessDistributionByPluginNameReq describes the process statistics
// request.
export interface GetProcessDistributionByPluginNameReq {
  exact_include_conditions: ProcessExactConditions;
  fuzzy_include_conditions: ProcessFuzzyConditions;
}

// GetProcessDistributionByPluginNameResp describes the process statistics
// response.
export interface GetProcessDistributionByPluginNameResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: Record<string, number>;
}

// DistinctProcessReq describes the process distinct request.
export interface DistinctProcessReq {
  exact_include_conditions: ProcessExactConditions;
  fuzzy_include_conditions: ProcessFuzzyConditions;
  // plugin_name serves the distinct from the in-memory cache keyed by
  // (tenant, plugin_name). When set, the conditions are ignored.
  plugin_name?: string;
}

// DistinctProcessResp describes the process distinct response.
export interface DistinctProcessResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: DistinctProcessRespData;
}

export interface DistinctProcessRespData {
  os_type: string[];
  cpu_arch: string[];
  version: string[];
  status: string[];
  plugin_name: string[];
  plugin_group: string[];
  plugin_pkg_name: string[];
}

