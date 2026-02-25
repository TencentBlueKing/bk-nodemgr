import forge from 'node-forge';

import { CipherService } from '@/api/modules/cipher';

// 内部变量，仅在模块内共享
let cachedPublicKey = '';

export const encryptionTool = {
  /**
   * 1. 获取公钥的请求方法（在项目入口或 setup 时调用）
   * 负责请求 API 并格式化 PEM 字符串
   */
  async initPublicKey() {
    try {
      const res = await CipherService.GetRSAPublicKey({});
      const pk = res?.public_key;
      if (pk) {
        // 彻底清洗并格式化为 node-forge 喜欢的每行 64 字符
        const pureBase64 = pk
          .replace(/-----BEGIN PUBLIC KEY-----|-----END PUBLIC KEY-----/gi, '')
          .replace(/\s+/g, '');
        const matched = pureBase64.match(/.{1,64}/g);
        cachedPublicKey = `-----BEGIN PUBLIC KEY-----\n${matched?.join('\n')}\n-----END PUBLIC KEY-----`;
        return true;
      }
    } catch (error) {
      console.error('获取公钥失败:', error);
    }
    return false;
  },

  /**
   * 2. 同步加密方法（在 handlePreview 或提交时使用）
   * 不带 async，直接返回加密结果
   */
  encryptSync(plainText: string): string | false {
    if (!plainText || !cachedPublicKey) {
      console.warn('加密失败：明文为空或公钥未加载');
      return false;
    }

    try {
      const publicKey = forge.pki.publicKeyFromPem(cachedPublicKey);

      // 适配后端：RSA-OAEP, SHA-256, Label: com.example.crypto.rsa.v1
      const rsaCiphertext = publicKey.encrypt(plainText, 'RSA-OAEP', {
        md: forge.md.sha256.create(),
        mgf1: {
          md: forge.md.sha256.create(),
        },
        label: 'com.example.crypto.rsa.v1',
      });

      // 二进制转换逻辑
      const rsaBytes = forge.util.binary.raw.decode(rsaCiphertext);

      // 拼接版本号 0x01 (1字节) + RSA密文 (256字节)
      const finalBytes = new Uint8Array(1 + rsaBytes.length);
      finalBytes[0] = 0x01;
      finalBytes.set(rsaBytes, 1);

      // 转回 Base64
      const finalRaw = forge.util.binary.raw.encode(finalBytes);
      return forge.util.encode64(finalRaw);
    } catch (error) {
      console.error('RSA加密过程异常:', error);
      return false;
    }
  },
};
