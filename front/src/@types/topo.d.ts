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
  data: TopoNetworkAreaListRespData;
}

export interface TopoNetworkAreaListRespData {
  total: number;
  items: NetworkArea[];
}

// TopoNetworkAreaStaticsReq describes the HTTP request body when query
// network-area statics in topo service.
export interface TopoNetworkAreaStaticsReq {
  bk_networkarea_id: number[];
}

// TopoNetworkAreaStaticsResp describes the HTTP response body when query
// network-area statics in topo service.
export interface TopoNetworkAreaStaticsResp {
  code: number;
  message: string;
  request_id: string;
  data: TopoNetworkAreaStaticsRespData;
}

export interface TopoNetworkAreaStaticsRespStaticsInfo {
  bk_networkarea_id: number;
  networkunit_count: number;
  proxy_count: number;
  agent_count: number;
}

export interface TopoNetworkAreaStaticsRespData {
  items: StaticsInfo[];
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
}

// TopoNetworkUnitListResp describes the HTTP response body when list
// network-unit in topo service.
export interface TopoNetworkUnitListResp {
  code: number;
  message: string;
  request_id: string;
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
}

// TopoNetworkUnitCreateResp describes the HTTP response body when create
// network unit in topo service.
export interface TopoNetworkUnitCreateResp {
  code: number;
  message: string;
  request_id: string;
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
}

// TopoNetworkUnitUpdateResp describes the HTTP response body when update
// network unit in topo service.
export interface TopoNetworkUnitUpdateResp {
  code: number;
  message: string;
  request_id: string;
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
  data: TopoHostListRespData;
}

export interface TopoHostListRespData {
  total: number;
  items: Host[];
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

// TopoEventExactConditions describes the conditions when list event
export interface TopoEventExactConditions {
  bk_networkarea_id: number[];
  bk_networkunit_id: number[];
  accesspoint_id: number[];
  type: string[];
  operator: string[];
}

// TopoEventListReq describes the HTTP request body when get host in topo
// service.
export interface TopoEventListReq {
  page: Page;
  only_count: boolean;
  exact_include_conditions: TopoEventExactConditions;
  operate_time_range: TimeRange;
}

// TopoEventListResp describes the HTTP response body when get host in topo
// service.
export interface TopoEventListResp {
  code: number;
  message: string;
  request_id: string;
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
  operate_time_range: TimeRange;
}

// TopoEventDistinctResp describes the HTTP response body when distinct
// topoevent in topo service.
export interface TopoEventDistinctResp {
  code: number;
  message: string;
  request_id: string;
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
  data: TopoConstantGetRespData;
}

export interface TopoConstantGetRespData {
  cloud_vendor: string[];
  os_type: string[];
}

