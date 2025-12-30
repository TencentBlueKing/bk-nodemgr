// gen-api.js 自动生成，请勿手动修改
import { type Config } from '../interceptors';

import type { GetRSAPublicKeyReq, GetRSAPublicKeyResp } from '@/@types/cipher';
import Fetch from '@/api/fetch';

const fetch = new Fetch({
  prefix: `${import.meta.env.BK_API_PREFIX}`,
});

export const CipherService = {
  // GetRSAPublicKey get rsa public key.
  GetRSAPublicKey: async <Request = GetRSAPublicKeyReq, ResponseData = GetRSAPublicKeyResp['data']>(params?: Request, config?: Config) => await fetch.post<Request, ResponseData>('/api/v3/cipher/rsa/get_public_key')(params, config),
};

