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
  cloud_vendor: string;
}

// Link describes the network unit link points target.
interface Link {
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  accesspoint_id: number;
}

// Links describes the network unit links.
interface Links {
  cluster: Link;
  file: Link;
  data: Link;
}

// Endpoints describes the links real address.
interface Endpoints {
  cluster: string[];
  file: string[];
  data: string[];
}

// InstallerRuntime describes the installer runtime config.
interface InstallerRuntime {
  base_work_dir: string;
}

// NodeRuntime describes the node runtime config.
interface NodeRuntime {
  base_deploy_dir: string;
  data_ipc: string;
  plugin_ipc: string;
  log_dir: string;
  zone_id: string;
  city_id: string;
}

// PluginRuntime describes the custom plugin runtime config.
interface PluginRuntime {
  base_deploy_dir: string;
  log_dir: string;
}

// CustomDeployConfig describes the deploy config for host deployment.
interface CustomDeployConfig {
  installer_runtime: InstallerRuntime;
  node_runtime: NodeRuntime;
  plugin_runtime: PluginRuntime;
}

// AccessPoint describes the access point informations.
interface AccessPoint {
  tenant_id: string;
  accesspoint_id: number;
  accesspoint_name: string;
  bk_networkarea_id: number;
  endpoints: Endpoints;
}

// AccessPointBrief describes the brief access point (no endpoints for
// sensitivity).
interface AccessPointBrief {
  accesspoint_id: number;
  accesspoint_name: string;
  bk_networkarea_id: number;
}

// NetworkUnit describes the network unit informations.
interface NetworkUnit {
  tenant_id: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  accesspoints: number[];
  links: Links;
  is_direct: boolean;
  direct_endpoints: Endpoints;
  generation: number;
  custom_deploy_config: Record<string, CustomDeployConfig>;
}

// NetworkUnitDetail describes the network unit detail informations.
// contains full access point objects.
interface NetworkUnitDetail {
  tenant_id: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  accesspoints: AccessPoint[];
  links: Links;
  is_direct: boolean;
  direct_endpoints: Endpoints;
  generation: number;
  custom_deploy_config: Record<string, CustomDeployConfig>;
}

// NetworkUnitBrief describes the network unit brief informations.
// only contains the access point ids.
interface NetworkUnitBrief {
  tenant_id: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  accesspoints: number[];
  links: Links;
  is_direct: boolean;
  generation: number;
}

// HostState describes the host state informations. Usually contains
// agent-related things.
interface HostState {
  node_role: string;
  node_status: string;
  node_version: string;
  bk_agent_id: string;
  node_generation: number;
  proxy_tags: string[];
}

// HostInfo describes the host info informations. Usually contains static
// configs.
interface HostInfo {
  bk_biz_id: number;
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  bk_host_name: string;
  dept_name: string;
  bk_host_innerip_list: string[];
  bk_host_innerip_v6_list: string[];
  bk_host_outerip_list: string[];
  bk_host_outerip_v6_list: string[];
  bk_mac: string;
  os_type: string;
  cpu_arch: string;
  bk_networkarea_name: string;
  bk_networkunit_name: string;
  login_ip: string;
  login_port: number;
  login_user: string;
  login_mode: string;
  login_credit_valid: boolean;
  export_ip: string;
  export_ip_v6: string;
  advertise_ip: string;
  advertise_ip_v6: string;
  relay_callback_port: number;
  relay_download_port: number;
  bk_addressing: string;
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

// NetworkUnitGraph describes the graph networkunit informations.
interface NetworkUnitGraph {
  tenant_id: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  accesspoints: number[];
}

// LinkGraph describes the graph link informations between networkunits.
interface LinkGraph {
  source_networkunit_id: number;
  target_networkunit_id: number;
  target_accesspoint_id: number;
  channel: string[];
}

// TopoEvent describes the topo event.
interface TopoEvent {
  tenant_id: string;
  type: string;
  bk_networkarea_id: number;
  bk_networkarea_name: string;
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  accesspoint_id: number;
  accesspoint_name: string;
  operate_time: number;
  operator: string;
}

// TimeRange describes the time range.
interface TimeRange {
  start_timestamp_sec: number;
  end_timestamp_sec: number;
}

// Release describes the release.
interface Release {
  name: string;
  generation: number;
  release_type: string;
  os_type: string;
  cpu_arch: string;
  version: string;
  file_name: string;
  labels: string[];
  enabled: boolean;
  as_default: boolean;
  md5: string;
  updated_at: number;
  operator: string;
  is_hidden: boolean;
}

interface ReleaseAgent {
  release: Release;
  change_log_en: string;
  change_log_zh: string;
}

interface ReleaseAgentBrief {
  generation: number;
  os_type: string;
  cpu_arch: string;
  version: string;
  enabled: boolean;
  as_default: boolean;
  is_hidden: boolean;
}

interface ReleaseProxy {
  release: Release;
  change_log_en: string;
  change_log_zh: string;
}

// ReleaseProxyBrief describes the brief release proxy (only core fields for
// list display).
interface ReleaseProxyBrief {
  generation: number;
  os_type: string;
  cpu_arch: string;
  version: string;
  enabled: boolean;
  as_default: boolean;
  change_log_en: string;
  change_log_zh: string;
  is_hidden: boolean;
}

// ReleasePlugin describes the release plugin.
interface ReleasePlugin {
  release: Release;
}

// ReleasePluginBrief describes the brief release plugin (only core fields for
// list display).
interface ReleasePluginBrief {
  generation: number;
  os_type: string;
  cpu_arch: string;
  version: string;
  enabled: boolean;
  as_default: boolean;
  is_hidden: boolean;
}

// ReleaseCert describes the release cert.
interface ReleaseCert {
  release: Release;
}

// ReleaseBinTool describes the release bin tool.
interface ReleaseBinTool {
  release: Release;
}

// ReleasePluginBinTool describes the release plugin bin tool.
interface ReleasePluginBinTool {
  release: Release;
}

// Platform describes the platform informations.
interface Platform {
  os_type: string;
  cpu_arch: string;
}

interface TargetVersion {
  version: string;
  cpu_arch: string;
  os_type: string;
}

// PackageEvent describes the package event.
interface PackageEvent {
  name: string;
  event_type: string;
  generation: number;
  release_type: string;
  os_type: string;
  cpu_arch: string;
  version: string;
  operate_time: number;
  operator: string;
}

// ConfigPolicyScope describes the config policy scope.
interface ConfigPolicyScope {
  bk_networkarea_id: number;
  bk_networkunit_id: number;
  os_type: string;
  cpu_arch: string;
}

// ConfigPolicyConfig describes the config policy config.
interface ConfigPolicyConfigItem {
  id: string;
  enabled: boolean;
  name_en: string;
  name_zh: string;
  remark_en: string;
  remark_zh: string;
  key: string;
  type: number;
  value_string: string;
  value_int: number;
  value_bool: boolean;
  value_string_select: string[];
  value_int_select: number[];
}

// ConfigPolicyConfigBlock describes the config policy config block.
interface ConfigPolicyConfigBlock {
  id: string;
  title_en: string;
  title_zh: string;
  items: ConfigPolicyConfigItem[];
}

// ConfigPolocy decribes the config policy.
interface ConfigPolicy {
  tenant_id: string;
  configpolicy_id: number;
  configpolicy_name: string;
  type: string;
  bk_biz_id: number;
  remark: string;
  scopes: ConfigPolicyScope[];
  configs: ConfigPolicyConfigBlock[];
  enabled: boolean;
  updated_time: number;
  operator: string;
  version: number;
  priority: number;
  target_host_ids: number[];
  target_plugin_name: string;
}

// ConfigPolicyEvent describes the config policy event.
interface ConfigPolicyEvent {
  tenant_id: string;
  configpolicy_id: number;
  configpolicy_name: string;
  configpolicy_type: string;
  type: string;
  version: number;
  operate_time: number;
  operator: string;
  bk_biz_id: number;
}

// ProcessInfo describes the process information.
interface ProcessInfo {
  pid: number;
  version: string;
  agent_id: string;
  auto_start: boolean;
  status: string;
}

// ProcessIdentity describes the process identity.
interface ProcessIdentity {
  name: string;
  setup_path: string;
  pid_path: string;
  config_path: string;
  log_path: string;
  user: string;
}

// ProcessController describes the process controller.
interface ProcessController {
  start_cmd: string;
  stop_cmd: string;
  restart_cmd: string;
  reload_cmd: string;
  kill_cmd: string;
  version_cmd: string;
  health_cmd: string;
}

// ProcessResource describes the process resource.
interface ProcessResource {
  cpu_limit_percent: number;
  mem_limit_percent: number;
}

// ProcessMonitorPolicy describes the process monitor policy.
interface ProcessMonitorPolicy {
  restart_type: string;
  start_check_seconds: number;
  stop_check_seconds: number;
  operate_timeout_seconds: number;
}

interface Process {
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
  bk_biz_id: number;
}

interface Plugin {
  tenant_id: string;
  name: string;
  group: string;
  pkg_name: string;
  memo: string;
  visible_biz_ids: number[];
}

interface Error {
  system: string;
  message: string;
  details: Details[];
}

interface ErrorDetails {
  code: string;
  message: string;
}

interface ResourceNode {
  type: string;
  type_name: string;
  id: string;
  name: string;
}

interface RelatedResourceType {
  system_id: string;
  system_name: string;
  type: string;
  type_name: string;
  instances: ResourceNode[];
}

interface Action {
  id: string;
  name: string;
  related_resource_types: RelatedResourceType[];
}

interface Permission {
  system: string;
  system_name: string;
  apply_url: string;
  actions: Action[];
}

// HostOpsInfo describes the out-of-band (ops) information for a host.
interface HostOpsInfo {
  bk_host_id: number;
  ops_console_host_id: number;
  ops_out_band_type: string;
  ops_out_band_protocol: string;
  ops_bmc_ip: string;
  ops_bmc_port: number;
}

