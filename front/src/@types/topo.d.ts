// gen-api.js 自动生成，请勿手动修改
// TopoBusinessListReq describes the HTTP request body when list business in
// topo service.
export interface TopoBusinessListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: TopoBusinessListReqExactConditions;
  fuzzy_include_conditions: TopoBusinessListReqFuzzyConditions;
}

export interface TopoBusinessListReqExactConditions {
  bk_biz_id: number[];
}

export interface TopoBusinessListReqFuzzyConditions {
  bk_biz_name: string[];
}

// TopoBusinessListResp describes the HTTP response body when list business in
// topo service.
export interface TopoBusinessListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoBusinessListRespData;
}

export interface TopoBusinessListRespData {
  total: number;
  items: Business[];
}

// TopoNetworkAreaListReq describes the HTTP request body when list network-area
// in topo service.
export interface TopoNetworkAreaListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: TopoNetworkAreaListReqExactConditions;
  fuzzy_include_conditions: TopoNetworkAreaListReqFuzzyConditions;
}

export interface TopoNetworkAreaListReqExactConditions {
  bk_networkarea_id: number[];
  cloud_vendor: string[];
}

export interface TopoNetworkAreaListReqFuzzyConditions {
  bk_networkarea_name: string[];
}

// TopoNetworkAreaListResp describes the HTTP response body when list
// network-area in topo service.
export interface TopoNetworkAreaListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkAreaListRespData;
}

export interface TopoNetworkAreaListRespData {
  total: number;
  items: NetworkArea[];
}

// TopoNetworkAreaStatisticsReq describes the HTTP request body when query
// network-area statics in topo service.
export interface TopoNetworkAreaStatisticsReq {
  bk_networkarea_id: number[];
}

// TopoNetworkAreaStatisticsResp describes the HTTP response body when query
// network-area statics in topo service.
export interface TopoNetworkAreaStatisticsResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkAreaStatisticsRespData;
}

export interface TopoNetworkAreaStatisticsRespStatisticsInfo {
  bk_networkarea_id: number;
  networkunit_count: number;
  proxy_count: number;
  agent_count: number;
  last_operator: string;
  last_operate_time: number;
}

export interface TopoNetworkAreaStatisticsRespData {
  items: StatisticsInfo[];
}

// TopoNetworkAreaGetReq describes the HTTP request body when get network-area
// in topo service.
export interface TopoNetworkAreaGetReq {
  bk_networkarea_id: number;
}

// TopoNetworkAreaGetResp describe the HTTP response body when get network-area
// in topo service.
export interface TopoNetworkAreaGetResp {
  code: number;
  message: string;
  request_id: string;
  data: NetworkArea;
}

// TopoNetworkAreaCreateReq describes the HTTP request body when create
// network-area in topo service.
export interface TopoNetworkAreaCreateReq {
  bk_networkarea_name: string;
  cloud_vendor: string;
}

// TopoNetworkAreaCreateResp describes the HTTP response body when create
// network-area in topo service.
export interface TopoNetworkAreaCreateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkAreaCreateRespData;
}

export interface TopoNetworkAreaCreateRespData {
  bk_networkarea_id: number;
}

// TopoNetworkAreaUpdateReq describes the HTTP request body when update
// network-area in topo service.
export interface TopoNetworkAreaUpdateReq {
  bk_networkarea_id: number;
  bk_networkarea_name: string;
  cloud_vendor: string;
}

// TopoNetworkAreaUpdateResp describes the HTTP response body when update
// network-area in topo service.
export interface TopoNetworkAreaUpdateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkAreaUpdateRespData;
}

export interface TopoNetworkAreaUpdateRespData {
  bk_networkarea_id: number;
}

// TopoNetworkAreaDeleteReq describes the HTTP request body when delete
// network-area in topo service.
export interface TopoNetworkAreaDeleteReq {
  bk_networkarea_id: number;
}

// TopoNetworkAreaDeleteResp describes the HTTP response body when delete
// network-area in topo service.
export interface TopoNetworkAreaDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkAreaDeleteRespData;
}

export interface TopoNetworkAreaDeleteRespData {
  bk_networkarea_id: number;
}

// TopoNetworkUnitListReq describes the HTTP request body when list network-unit
// in topo service.
export interface TopoNetworkUnitListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: TopoNetworkUnitListReqExactConditions;
}

export interface TopoNetworkUnitListReqExactConditions {
  bk_networkunit_id: number[];
  bk_networkarea_id: number[];
  is_direct: boolean[];
}

// TopoNetworkUnitListResp describes the HTTP response body when list
// network-unit in topo service.
export interface TopoNetworkUnitListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkUnitListRespData;
}

export interface TopoNetworkUnitListRespData {
  total: number;
  items: NetworkUnit[];
}

// TopoNetworkUnitGetReq describes the HTTP request body when get network-unit
// in topo service.
export interface TopoNetworkUnitGetReq {
  bk_networkunit_id: number;
}

// TopoNetworkUnitGetResp describe the HTTP response body when get network-unit
// in topo service.
export interface TopoNetworkUnitGetResp {
  code: number;
  message: string;
  request_id: string;
  data: NetworkUnit;
}

// TopoNetworkUnitCreateReq describes the HTTP request body when create
// network-unit in topo service.
export interface TopoNetworkUnitCreateReq {
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  accesspoints: AccessPoint[];
  links: Links;
  is_direct: boolean;
  direct_endpoints: Endpoints;
}

// TopoNetworkUnitCreateResp describes the HTTP response body when create
// network unit in topo service.
export interface TopoNetworkUnitCreateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkUnitCreateRespData;
}

export interface TopoNetworkUnitCreateRespData {
  bk_networkunit_id: number;
}

// TopoNetworkUnitUpdateReq describes the HTTP request body when update
// network-unit in topo service.
export interface TopoNetworkUnitUpdateReq {
  bk_networkunit_id: number;
  bk_networkunit_name: string;
  bk_networkarea_id: number;
  accesspoints: AccessPoint[];
  links: Links;
  is_direct: boolean;
  direct_endpoints: Endpoints;
}

// TopoNetworkUnitUpdateResp describes the HTTP response body when update
// network unit in topo service.
export interface TopoNetworkUnitUpdateResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkUnitUpdateRespData;
}

export interface TopoNetworkUnitUpdateRespData {
  bk_networkunit_id: number;
}

// TopoNetworkUnitDeleteReq describes the HTTP request body when delete
// network unit in topo service.
export interface TopoNetworkUnitDeleteReq {
  bk_networkunit_id: number;
}

// TopoNetworkUnitDeleteResp describes the HTTP response body when delete
// network unit in topo service.
export interface TopoNetworkUnitDeleteResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoNetworkUnitDeleteRespData;
}

export interface TopoNetworkUnitDeleteRespData {
  bk_networkunit_id: number;
}

// TopoHostExactConditions describes host exact conditions.
export interface TopoHostExactConditions {
  bk_host_id: number[];
  bk_biz_id: number[];
  bk_networkarea_id: number[];
  os_type: string[];
  node_role: string[];
  node_status: string[];
  node_version: string[];
  bk_agent_id: string[];
  bk_networkunit_id: number[];
  node_generation: number[];
}

// TopoHostFuzzyConditions describes host fuzzy conditions.
export interface TopoHostFuzzyConditions {
  bk_host_name: string[];
  dept_name: string[];
  bk_host_innerip: string[];
  bk_host_innerip_v6: string[];
  bk_host_outerip: string[];
  bk_host_outerip_v6: string[];
}

// TopoHostListReq describes the HTTP request body when list host in topo
// service.
export interface TopoHostListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: TopoHostExactConditions;
  fuzzy_include_conditions: TopoHostFuzzyConditions;
}

// TopoHostListResp describes the HTTP response body when list host in topo
// service.
export interface TopoHostListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoHostListRespData;
}

export interface TopoHostListRespData {
  total: number;
  items: Host[];
}

// TopoHostSelectHostIDReq describes the HTTP request body when select host
// id in topo service.
export interface TopoHostSelectHostIDReq {
  exact_include_conditions: TopoHostExactConditions;
  fuzzy_include_conditions: TopoHostFuzzyConditions;
  exact_exclude_conditions: TopoHostExactConditions;
}

// TopoHostSelectHostIDResp  describes the HTTP response body when select
// host id in topo service.
export interface TopoHostSelectHostIDResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoHostSelectHostIDRespData;
}

export interface TopoHostSelectHostIDRespData {
  items: number[];
}

// TopoHostSelectInnerIPReq describes the HTTP request body when select host
// inner ip in topo service.
export interface TopoHostSelectInnerIPReq {
  exact_include_conditions: TopoHostExactConditions;
  fuzzy_include_conditions: TopoHostFuzzyConditions;
  exact_exclude_conditions: TopoHostExactConditions;
}

// TopoHostSelectInnerIPResp  describes the HTTP response body when select
// host inner ip in topo service.
export interface TopoHostSelectInnerIPResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoHostSelectInnerIPRespData;
}

export interface TopoHostSelectInnerIPRespData {
  items: string[];
}

// TopoHostSelectInnerIPV6Req describes the HTTP request body when select host
// inner ipv6 in topo service.
export interface TopoHostSelectInnerIPV6Req {
  exact_include_conditions: TopoHostExactConditions;
  fuzzy_include_conditions: TopoHostFuzzyConditions;
  exact_exclude_conditions: TopoHostExactConditions;
}

// TopoHostSelectInnerIPV6Resp  describes the HTTP response body when select
// host inner ipv6 in topo service.
export interface TopoHostSelectInnerIPV6Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoHostSelectInnerIPV6RespData;
}

export interface TopoHostSelectInnerIPV6RespData {
  items: string[];
}

// TopoHostSelectNetWorkareaIDAndInnerIPReq describes the HTTP request body when
// select host networkarea id and inner ip in topo service.
export interface TopoHostSelectNetWorkareaIDAndInnerIPReq {
  exact_include_conditions: TopoHostExactConditions;
  fuzzy_include_conditions: TopoHostFuzzyConditions;
  exact_exclude_conditions: TopoHostExactConditions;
}

// TopoHostSelectNetWorkareaIDAndInnerIPResp  describes the HTTP response body
// when select host networkarea id and inner ip in topo service.
export interface TopoHostSelectNetWorkareaIDAndInnerIPResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoHostSelectNetWorkareaIDAndInnerIPRespData;
}

export interface TopoHostSelectNetWorkareaIDAndInnerIPRespData {
  items: string[];
}

// TopoHostSelectNetWorkareaIDAndInnerIPV6Req describes the HTTP request body
// when select host networkarea id and inner ipv6 in topo service.
export interface TopoHostSelectNetWorkareaIDAndInnerIPV6Req {
  exact_include_conditions: TopoHostExactConditions;
  fuzzy_include_conditions: TopoHostFuzzyConditions;
  exact_exclude_conditions: TopoHostExactConditions;
}

// TopoHostSelectNetWorkareaIDAndInnerIPV6Resp  describes the HTTP response body
// when select host networkarea id and inner ipv6 in topo service.
export interface TopoHostSelectNetWorkareaIDAndInnerIPV6Resp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoHostSelectNetWorkareaIDAndInnerIPV6RespData;
}

export interface TopoHostSelectNetWorkareaIDAndInnerIPV6RespData {
  items: string[];
}

// TopoHostDistinctReq describes the HTTP request body when distinct host in
// topp service.
export interface TopoHostDistinctReq {
  exact_include_conditions: TopoHostExactConditions;
  fuzzy_include_conditions: TopoHostFuzzyConditions;
}

// TopoHostDistinctResp describes the HTTP response body when distinct host in
// topp service.
export interface TopoHostDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoHostDistinctRespData;
}

export interface TopoHostDistinctRespData {
  node_role: string[];
  node_status: string[];
  node_version: string[];
  dept_name: string[];
  os_type: string[];
  arch: string[];
  addressing: string[];
  bk_networkarea_id: number[];
  bk_networkunit_id: number[];
  bk_biz_id: number[];
}

// TopoGraphGetReq describes the HTTP request body when get graph in topo
// service.
export interface TopoGraphGetReq {
  bk_networkarea_id: number[];
}

// TopoGraphGetResp describes the HTTP response body when get graph in topo
// service.
export interface TopoGraphGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoGraphGetRespData;
}

export interface TopoGraphGetRespData {
  networkunit: NetworkUnitGraph[];
  links: LinkGraph[];
}

// TopoGraphNodeCountReq describes the HTTP request body when get node count
export interface TopoGraphNodeCountReq {
  bk_networkunit_id: number[];
}

// TopoGraphNodeCountResp describes the HTTP response body when get node count
export interface TopoGraphNodeCountResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoGraphNodeCountRespData;
}

export interface TopoGraphNodeCountRespNodeInfo {
  bk_networkunit_id: number;
  proxy: number;
  agent: number;
}

export interface TopoGraphNodeCountRespData {
  networkunits: NodeInfo[];
}

// TopoGraphNodeGetReq describes the HTTP request body when get node
export interface TopoGraphNodeGetReq {
  bk_networkunit_id: number[];
}

// TopoGraphNodeGetResp describes the HTTP response body when get node
export interface TopoGraphNodeGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoGraphNodeGetRespData;
}

export interface TopoGraphNodeGetRespGraphNodeInfo {
  bk_networkunit_id: number;
  running_proxy: number;
  total_proxy: number;
  running_agent: number;
  total_agent: number;
  is_healthy: boolean;
  cycle_times: string[];
}

export interface TopoGraphNodeGetRespData {
  graph_node_info: GraphNodeInfo[];
}

// TopoEventExactConditions describes the conditions when list event
export interface TopoEventExactConditions {
  bk_networkarea_id: number[];
  bk_networkunit_id: number[];
  accesspoint_id: number[];
  type: string[];
  operator: string[];
}

// TopoEventFuzzyConditions describes the conditions when list event
export interface TopoEventFuzzyConditions {
  bk_networkarea_name: string[];
  bk_networkunit_name: string[];
}

// TopoEventListReq describes the HTTP request body when get host in topo
// service.
export interface TopoEventListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: TopoEventExactConditions;
  fuzzy_include_conditions: TopoEventFuzzyConditions;
  operate_time_range: TimeRange;
}

// TopoEventListResp describes the HTTP response body when get host in topo
// service.
export interface TopoEventListResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoEventListRespData;
}

export interface TopoEventListRespData {
  total: number;
  items: TopoEvent[];
}

// TopoEventDistinctReq describes the HTTP request body when distinct topoevent
// in topo service.
export interface TopoEventDistinctReq {
  exact_include_conditions: TopoEventExactConditions;
  fuzzy_include_conditions: TopoEventFuzzyConditions;
  operate_time_range: TimeRange;
}

// TopoEventDistinctResp describes the HTTP response body when distinct
// topoevent in topo service.
export interface TopoEventDistinctResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoEventDistinctRespData;
}

export interface TopoEventDistinctRespData {
  bk_networkarea_id: number[];
  bk_networkunit_id: number[];
  accesspoint_id: number[];
  type: string[];
  operator: string[];
}

// TopoConstantGetReq describes the HTTP request body when get constant in topo
// service.
export interface TopoConstantGetReq {
  cloud_vendor: boolean;
  os_type: boolean;
}

// TopoConstantGetResp describes the HTTP response body when get constant in
// topo service.
export interface TopoConstantGetResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: TopoConstantGetRespData;
}

export interface TopoConstantGetRespData {
  cloud_vendor: string[];
  os_type: string[];
}

