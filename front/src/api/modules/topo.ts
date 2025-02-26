// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { TopoBusinessListReq, TopoBusinessListResp, TopoNetworkAreaListReq, TopoNetworkAreaListResp, TopoNetworkAreaCreateReq, TopoNetworkAreaCreateResp, TopoNetworkAreaGetReq, TopoNetworkAreaGetResp, TopoNetworkAreaUpdateReq, TopoNetworkAreaUpdateResp, TopoNetworkAreaDeleteReq, TopoNetworkAreaDeleteResp, TopoNetworkUnitListReq, TopoNetworkUnitListResp, TopoNetworkUnitCreateReq, TopoNetworkUnitCreateResp, TopoNetworkUnitGetReq, TopoNetworkUnitGetResp, TopoNetworkUnitUpdateReq, TopoNetworkUnitUpdateResp, TopoNetworkUnitDeleteReq, TopoNetworkUnitDeleteResp, TopoHostListReq, TopoHostListResp } from '@/@types/topo';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const TopoService = {
  // BusinessList provides business listing.
  BusinessList: async <Request = TopoBusinessListReq, ResponseData = TopoBusinessListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/business/list')(params, config),
  // NetworkAreaList provides network-area listing.
  NetworkAreaList: async <Request = TopoNetworkAreaListReq, ResponseData = TopoNetworkAreaListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkarea/list')(params, config),
  // NetworkAreaCreate provides creating a new network-area.
  NetworkAreaCreate: async <Request = TopoNetworkAreaCreateReq, ResponseData = TopoNetworkAreaCreateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkarea/create')(params, config),
  // NetworkAreaGet provides getting an existing network-area by id.
  NetworkAreaGet: async <Request = TopoNetworkAreaGetReq, ResponseData = TopoNetworkAreaGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkarea/get')(params, config),
  // NetworkAreaUpdate provides updating an existing network-area by id.
  NetworkAreaUpdate: async <Request = TopoNetworkAreaUpdateReq, ResponseData = TopoNetworkAreaUpdateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkarea/update')(params, config),
  // NetworkAreaDelete provides deleting an existing network-area by id.
  NetworkAreaDelete: async <Request = TopoNetworkAreaDeleteReq, ResponseData = TopoNetworkAreaDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkarea/delete')(params, config),
  // NetworkUnitList provides network-unit listing.
  NetworkUnitList: async <Request = TopoNetworkUnitListReq, ResponseData = TopoNetworkUnitListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkunit/list')(params, config),
  // NetworkUnitCreate provides creating a new network-unit.
  NetworkUnitCreate: async <Request = TopoNetworkUnitCreateReq, ResponseData = TopoNetworkUnitCreateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkunit/create')(params, config),
  // NetworkUnitGet provides getting an existing network-unit by id.
  NetworkUnitGet: async <Request = TopoNetworkUnitGetReq, ResponseData = TopoNetworkUnitGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkunit/get')(params, config),
  // NetworkUnitUpdate provides updating an existing network-unit by id.
  NetworkUnitUpdate: async <Request = TopoNetworkUnitUpdateReq, ResponseData = TopoNetworkUnitUpdateResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkunit/update')(params, config),
  // NetworkUnitDelete provides deleting an existing network-unit by id.
  NetworkUnitDelete: async <Request = TopoNetworkUnitDeleteReq, ResponseData = TopoNetworkUnitDeleteResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkunit/delete')(params, config),
  // HostList provides business listing.
  HostList: async <Request = TopoHostListReq, ResponseData = TopoHostListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/host/list')(params, config),
};

