// gen-api.js 自动生成，请勿手动修改
// Backend 实例启动时加载的业务配置白名单，不包含凭据、连接端点及内部观测参数。
export interface SystemConfig {
  // 租户模式：single 或 multiple。
  tenant_mode: string;
  system: SystemConfigSystem;
  iam: SystemConfigIAM;
  node_event_data_id: SystemConfigNodeEventDataID;
  default_direct_network_unit: SystemConfigDefaultDirectNetworkUnit;
  gse_deploy_configs: SystemConfigGSEDeployConfig[];
}

// 部署环境及版本形态。
export interface SystemConfigSystem {
  env: string;
  edition: string;
}

// IAM 权限管理开关。
export interface SystemConfigIAM {
  v3_enabled: boolean;
  v4_enabled: boolean;
}

// 全局默认节点事件 Data ID，不解析业务级覆盖。
export interface SystemConfigNodeEventDataID {
  agent_base_alarm_event_data_id: number;
  task_proc_event_data_id: number;
}

// 默认直连网络单元的自动创建设置，不表示已存在的网络单元状态。
export interface SystemConfigDefaultDirectNetworkUnit {
  enabled: boolean;
  name: string;
}

// 按 GSE
// 代际和操作系统配置的安装默认设置，不解析单次部署覆盖或目标主机实际状态。
export interface SystemConfigGSEDeployConfig {
  generation: number;
  os_type: string;
  base_work_dir: string;
  base_deploy_dir: string;
  // Agent 扩展配置目录；为空表示未配置覆盖值。
  extra_config_dir: string;
  // 插件读取的主机 ID 文件路径；为空表示未配置覆盖值。
  host_id_path: string;
  // 插件 V2 兼容安装的运行目录；为空表示未配置覆盖值。
  v2_run_dir: string;
  // 插件 V2 兼容安装的数据目录；为空表示未配置覆盖值。
  v2_data_dir: string;
  // 节点部署的地域标识；为空表示未配置覆盖值。
  zone_id: string;
  // 节点部署的城市标识；为空表示未配置覆盖值。
  city_id: string;
  // 安装 HTTP 代理策略：inherit 沿用主机环境代理，direct 强制直连，specified
  // 使用指定代理；不返回代理地址。
  http_proxy_mode: string;
  // Windows
  // 安装器管理的防火墙功能：gse、nodemgr；空数组表示不管理规则。不返回允许访问的
  // IP/CIDR。
  firewall_features: string[];
}

// 查询 Backend 业务配置，无筛选参数。
export interface GetSystemConfigReq {
}

// Backend 业务配置只读视图。
export interface GetSystemConfigResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: SystemConfig;
  permission: Permission;
}

