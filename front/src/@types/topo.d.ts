// gen-api.js 自动生成，请勿手动修改
// TopoBusinessListReq describes the HTTP request body when list business in
// topo service.
export interface TopoBusinessListReq {
  page: Page;
  onlyCount: boolean;
}

// TopoBusinessListResp describes the HTTP response body when list business in
// topo service.
export interface TopoBusinessListResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkAreaListReq describes the HTTP request body when list network-area
// in topo service.
export interface TopoNetworkAreaListReq {
  page: Page;
  onlyCount: boolean;
  includeConditions: Conditions;
}

// TopoNetworkAreaListResp describes the HTTP response body when list
// network-area in topo service.
export interface TopoNetworkAreaListResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkAreaGetReq describes the HTTP request body when get network-area
// in topo service.
export interface TopoNetworkAreaGetReq {
  bkNetworkareaId: number;
}

// TopoNetworkAreaGetResp describe the HTTP response body when get network-area
// in topo service.
export interface TopoNetworkAreaGetResp {
  code: number;
  message: string;
  requestId: string;
  data: NetworkArea;
}

// TopoNetworkAreaCreateReq describes the HTTP request body when create
// network-area in topo service.
export interface TopoNetworkAreaCreateReq {
  bkNetworkareaName: string;
  bkCloudVendor: number;
}

// TopoNetworkAreaCreateResp describes the HTTP response body when create
// network-area in topo service.
export interface TopoNetworkAreaCreateResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkAreaUpdateReq describes the HTTP request body when update
// network-area in topo service.
export interface TopoNetworkAreaUpdateReq {
  bkNetworkareaId: number;
  bkNetworkareaName: string;
  bkCloudVendor: number;
}

// TopoNetworkAreaUpdateResp describes the HTTP response body when update
// network-area in topo service.
export interface TopoNetworkAreaUpdateResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkAreaDeleteReq describes the HTTP request body when delete
// network-area in topo service.
export interface TopoNetworkAreaDeleteReq {
  bkNetworkareaId: number;
}

// TopoNetworkAreaDeleteResp describes the HTTP response body when delete
// network-area in topo service.
export interface TopoNetworkAreaDeleteResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkUnitListReq describes the HTTP request body when list network-unit
// in topo service.
export interface TopoNetworkUnitListReq {
  page: Page;
  onlyCount: boolean;
}

// TopoNetworkUnitListResp describes the HTTP response body when list
// network-unit in topo service.
export interface TopoNetworkUnitListResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkUnitGetReq describes the HTTP request body when get network-unit
// in topo service.
export interface TopoNetworkUnitGetReq {
  bkNetworkunitId: number;
}

// TopoNetworkUnitGetResp describe the HTTP response body when get network-unit
// in topo service.
export interface TopoNetworkUnitGetResp {
  code: number;
  message: string;
  requestId: string;
  data: NetworkUnit;
}

// TopoNetworkUnitCreateReq describes the HTTP request body when create
// network-unit in topo service.
export interface TopoNetworkUnitCreateReq {
  bkNetworkunitName: string;
  accessPoints: AccessPoint[];
  links: Links;
}

// TopoNetworkUnitCreateResp describes the HTTP response body when create
// network unit in topo service.
export interface TopoNetworkUnitCreateResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkUnitUpdateReq describes the HTTP request body when update
// network-unit in topo service.
export interface TopoNetworkUnitUpdateReq {
  bkNetworkunitId: number;
  bkNetworkunitName: string;
  accessPoints: AccessPoint[];
  links: Links;
}

// TopoNetworkUnitUpdateResp describes the HTTP response body when update
// network unit in topo service.
export interface TopoNetworkUnitUpdateResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoNetworkUnitDeleteReq describes the HTTP request body when delete
// network unit in topo service.
export interface TopoNetworkUnitDeleteReq {
  bkNetworkunitId: number;
}

// TopoNetworkUnitDeleteResp describes the HTTP response body when delete
// network unit in topo service.
export interface TopoNetworkUnitDeleteResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

// TopoHostListReq describes the HTTP request body when list host in topo
// service.
export interface TopoHostListReq {
  page: Page;
  onlyCount: boolean;
  includeConditions: Conditions;
}

// TopoHostListResp describes the HTTP response body when list host in topo
// service.
export interface TopoHostListResp {
  code: number;
  message: string;
  requestId: string;
  data: Data;
}

