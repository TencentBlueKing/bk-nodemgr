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
  permission: Permission;
  data: GetRSAPublicKeyRespData;
}

export interface GetRSAPublicKeyRespData {
  public_key: string;
}

// GetCurrentPublicKeyReq get current public key request message.
export interface GetCurrentPublicKeyReq {
}

// GetCurrentPublicKeyResp get current public key response message, the
// public key of the globally enabled credential encryption suite.
export interface GetCurrentPublicKeyResp {
  code: number;
  message: string;
  request_id: string;
  error: Error;
  permission: Permission;
  data: GetCurrentPublicKeyRespData;
}

export interface GetCurrentPublicKeyRespData {
  // key_type is the asymmetric algorithm of public_key: RSA4096 (CLASSIC
  // suite) or SM2 (SHANGMI suite).
  key_type: string;
  public_key: string;
}

