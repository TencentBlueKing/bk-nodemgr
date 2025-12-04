// gen-api.js 自动生成，请勿手动修改
// ProcessListReq describes the process list request.
export interface ProcessListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: ProcessListReqExactConditions;
  fuzzy_include_conditions: ProcessListReqFuzzyConditions;
}

export interface ProcessListReqExactConditions {
  bk_host_id: number[];
  plugin_group: string[];
  node_generation: string[];
  platform_os: string[];
  platform_arch: string[];
  status: string[];
  agent_id: string[];
  version: string[];
  plugin_name: string[];
  plugin_pkg_name: string[];
}

export interface ProcessListReqFuzzyConditions {
  name: string[];
  plugin_pkg_name: string[];
}

// ProcessInfo describes the process information.
export interface ProcessInfo {
  pid: number;
  version: string;
  agent_id: string;
  trusteeship: boolean;
  status: string;
}

// ProcessIdentity describes the process identity.
export interface ProcessIdentity {
  name: string;
  setup_path: string;
  pid_path: string;
  config_path: string;
  log_path: string;
  user: string;
}

// ProcessController describes the process controller.
export interface ProcessController {
  start_cmd: string;
  stop_cmd: string;
  restart_cmd: string;
  reload_cmd: string;
  kill_cmd: string;
  version_cmd: string;
  health_cmd: string;
}

// ProcessResource describes the process resource.
export interface ProcessResource {
  cpu_limit_percent: number;
  mem_limit_percent: number;
}

// ProcessMonitorPolicy describes the process monitor policy.
export interface ProcessMonitorPolicy {
  auto_type: string;
  start_check_seconds: number;
  stop_check_seconds: number;
  operate_timeout_seconds: number;
}

// ProcessListResp describes the process list response.
export interface ProcessListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: ProcessListRespData;
}

export interface ProcessListRespProcess {
  tenant_id: string;
  bk_host_id: number;
  plugin_name: string;
  plugin_pkg_name: string;
  plugin_group: string;
  platform: Platform;
  generation: number;
  process_info: ProcessInfo;
  process_identity: ProcessIdentity;
  process_controller: ProcessController;
  process_resource: ProcessResource;
  process_monitor_policy: ProcessMonitorPolicy;
}

export interface ProcessListRespData {
  total: number;
  items: Process[];
}

