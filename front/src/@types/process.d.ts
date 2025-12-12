// gen-api.js 自动生成，请勿手动修改
// ProcessListResp describes the process list response.
export interface ProcessListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
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
  data: Record<string, number>;
}

