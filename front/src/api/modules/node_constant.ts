// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { NodeConstantDeployGetReq, NodeConstantDeployGetResp } from '@/@types/node_constant';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const NodeConstantService = {
  // NodeConstantDeployGet provides getting deploy constant.
  NodeConstantDeployGet: async <Request = NodeConstantDeployGetReq, ResponseData = NodeConstantDeployGetResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/node/constant/deploy/get')(params, config),
};

