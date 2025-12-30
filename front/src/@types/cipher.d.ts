// gen-api.js 自动生成，请勿手动修改
// GetRSAPublicKeyReq get rsa public key request message.
export interface GetRSAPublicKeyReq {
}

// GetRSAPublicKeyResp get rsa public key response message.
export interface GetRSAPublicKeyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  data: GetRSAPublicKeyRespData;
}

export interface GetRSAPublicKeyRespData {
  public_key: string;
}

