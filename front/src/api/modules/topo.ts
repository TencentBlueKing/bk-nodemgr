// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { TopoBusinessListReq, TopoBusinessListResp, TopoNetworkAreaListReq, TopoNetworkAreaListResp, TopoNetworkAreaStatisticsReq, TopoNetworkAreaStatisticsResp, TopoNetworkAreaCreateReq, TopoNetworkAreaCreateResp, TopoNetworkAreaGetReq, TopoNetworkAreaGetResp, TopoNetworkAreaUpdateReq, TopoNetworkAreaUpdateResp, TopoNetworkAreaDeleteReq, TopoNetworkAreaDeleteResp, TopoNetworkUnitListReq, TopoNetworkUnitListResp, TopoNetworkUnitCreateReq, TopoNetworkUnitCreateResp, TopoNetworkUnitGetReq, TopoNetworkUnitGetResp, TopoNetworkUnitUpdateReq, TopoNetworkUnitUpdateResp, TopoNetworkUnitDeleteReq, TopoNetworkUnitDeleteResp, TopoHostListReq, TopoHostListResp, TopoHostDistinctReq, TopoHostDistinctResp, TopoGraphGetReq, TopoGraphGetResp, TopoGraphNodeCountReq, TopoGraphNodeCountResp, TopoEventListReq, TopoEventListResp, TopoEventDistinctReq, TopoEventDistinctResp, TopoConstantGetReq, TopoConstantGetResp, TopoGraphNodeGetResp } from '@/@types/topo';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const TopoService = {
  // BusinessList provides business listing.
  BusinessList: async <Request = TopoBusinessListReq, ResponseData = TopoBusinessListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/business/list')(params, config),
  // NetworkAreaList provides network-area listing.
  NetworkAreaList: async <Request = TopoNetworkAreaListReq, ResponseData = TopoNetworkAreaListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkarea/list')(params, config),
  // NetworkAreaStatistics provides network-area statistics.
  NetworkAreaStatistics: async <Request = TopoNetworkAreaStatisticsReq, ResponseData = TopoNetworkAreaStatisticsResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/networkarea/statistics')(params, config),
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
  // HostList provides host listing.
  HostList: async <Request = TopoHostListReq, ResponseData = TopoHostListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/host/list')(params, config),
  // HostDistinct provides host distincting.
  HostDistinct: async <Request = TopoHostDistinctReq, ResponseData = TopoHostDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/host/distinct')(params, config),
  // GraphGet provides getting graph nodes and edges.
  GraphGet: async <Request = TopoGraphGetReq, ResponseData = TopoGraphGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/graph/get')(params, config),
  // GraphNodeCount provides getting graph node count.
  GraphNodeCount: async <Request = TopoGraphNodeCountReq, ResponseData = TopoGraphNodeCountResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/graph/node/count')(params, config),
  // EventList provides event listing.
  EventList: async <Request = TopoEventListReq, ResponseData = TopoEventListResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/event/list')(params, config),
  // EventDistinct provides event distincting.
  EventDistinct: async <Request = TopoEventDistinctReq, ResponseData = TopoEventDistinctResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/event/distinct')(params, config),
  // ConstantGet provides getting a constant.
  ConstantGet: async <Request = TopoConstantGetReq, ResponseData = TopoConstantGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/constant/get')(params, config),
  // TopoGraphNodeGetReq provides graph node getting.
  TopoGraphNodeGetReq: async <Request = TopoGraphNodeGetResp, ResponseData = TopoGraphNodeGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/topo/graph/node/get')(params, config),
};

