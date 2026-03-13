import forge from 'node-forge';

import { CipherService } from '@/api/modules/cipher';

// ── 常量 ──────────────────────────────────────────────
const VERSION_V1 = 0x01; // RSA-OAEP 直接加密
const VERSION_V2 = 0x02; // AES-GCM + RSA-OAEP 混合加密
const RSA_OAEP_LABEL = 'com.example.crypto.rsa.v1';
const AES_KEY_BYTES = 32; // AES-256
const AES_NONCE_BYTES = 12; // GCM recommended nonce length
const AES_TAG_BYTES = 16; // GCM authentication tag length

// ── 内部变量，仅在模块内共享 ──────────────────────────
let cachedPublicKey = '';
/** RSA-OAEP with SHA-256 单块可加密的最大明文长度（字节） */
let maxRsaPlainBytes = 0;
/** 防止并发调用 initPublicKey 时重复请求 */
let pendingPromise: Promise<boolean> | null = null;

/**
 * 根据 RSA 公钥位数计算 RSA-OAEP/SHA-256 单块最大明文长度
 * 公式: keyBytes - 2 * hashBytes - 2
 * 对于 2048-bit key: 256 - 64 - 2 = 190
 */
function calcMaxPlainBytes(pubKey: forge.pki.rsa.PublicKey): number {
  const keyBytes = Math.ceil(pubKey.n.bitLength() / 8);
  const hashBytes = 32; // SHA-256 → 32 bytes
  return keyBytes - 2 * hashBytes - 2;
}

// ── RSA-OAEP 加密（v1/v2 共用） ──────────────────────
function rsaOaepEncrypt(pubKey: forge.pki.rsa.PublicKey, data: string): string {
  return pubKey.encrypt(data, 'RSA-OAEP', {
    md: forge.md.sha256.create(),
    mgf1: { md: forge.md.sha256.create() },
    label: RSA_OAEP_LABEL,
  });
}

// ── V1: 纯 RSA-OAEP ─────────────────────────────────
function encryptV1(pubKey: forge.pki.rsa.PublicKey, plainText: string): string {
  const rsaCiphertext = rsaOaepEncrypt(pubKey, plainText);
  const rsaBytes = forge.util.binary.raw.decode(rsaCiphertext);

  // 0x01 (1B) + rsaCiphertext
  const out = new Uint8Array(1 + rsaBytes.length);
  out[0] = VERSION_V1;
  out.set(rsaBytes, 1);

  return forge.util.encode64(forge.util.binary.raw.encode(out));
}

// ── V2: AES-256-GCM + RSA-OAEP 混合 ─────────────────
function encryptV2(pubKey: forge.pki.rsa.PublicKey, plainText: string): string {
  // 1. 生成随机 AES-256 key 与 12B nonce
  const aesKeyRaw = forge.random.getBytesSync(AES_KEY_BYTES);
  const nonce = forge.random.getBytesSync(AES_NONCE_BYTES);

  // 2. AES-GCM 加密明文
  const cipher = forge.cipher.createCipher('AES-GCM', aesKeyRaw);
  cipher.start({ iv: nonce, tagLength: AES_TAG_BYTES * 8, additionalData: RSA_OAEP_LABEL });
  cipher.update(forge.util.createBuffer(forge.util.encodeUtf8(plainText)));
  cipher.finish();

  const aesGcmCiphertext = cipher.output.getBytes(); // 密文（不含 tag）
  const tag = cipher.mode.tag.getBytes(); // 认证标签

  // 3. RSA-OAEP 包裹 AES key
  const wrappedKey = rsaOaepEncrypt(pubKey, aesKeyRaw);

  // 4. 组装二进制：0x02 (1B) + wrappedKey + nonce (12B) + aesGcmCiphertext + tag (16B)
  const wrappedKeyBytes = forge.util.binary.raw.decode(wrappedKey);
  const nonceBytes = forge.util.binary.raw.decode(nonce);
  const ciphertextBytes = forge.util.binary.raw.decode(aesGcmCiphertext);
  const tagBytes = forge.util.binary.raw.decode(tag);

  const totalLen = 1 + wrappedKeyBytes.length + nonceBytes.length + ciphertextBytes.length + tagBytes.length;
  const out = new Uint8Array(totalLen);
  let offset = 0;

  out[offset] = VERSION_V2;
  offset += 1;

  out.set(wrappedKeyBytes, offset);
  offset += wrappedKeyBytes.length;

  out.set(nonceBytes, offset);
  offset += nonceBytes.length;

  out.set(ciphertextBytes, offset);
  offset += ciphertextBytes.length;

  out.set(tagBytes, offset);

  return forge.util.encode64(forge.util.binary.raw.encode(out));
}

// ── 导出 ─────────────────────────────────────────────
export const encryptionTool = {
  /**
   * 1. 获取公钥的请求方法（在项目入口或 setup 时调用）
   * 负责请求 API 并格式化 PEM 字符串，同时计算 RSA 最大明文长度
   */
  async initPublicKey() {
    // Already initialized — skip network request
    if (cachedPublicKey) {
      return true;
    }
    // If a request is already in-flight, reuse it
    if (pendingPromise) {
      return pendingPromise;
    }
    pendingPromise = (async () => {
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

          // 预计算单块 RSA 最大可加密明文长度
          const pubKey = forge.pki.publicKeyFromPem(cachedPublicKey);
          maxRsaPlainBytes = calcMaxPlainBytes(pubKey as forge.pki.rsa.PublicKey);

          return true;
        }
      } catch (error) {
        console.error('获取公钥失败:', error);
      }
      return false;
    })().finally(() => {
      pendingPromise = null;
    });
    return pendingPromise;
  },

  /**
   * 2. 同步加密方法（在 handlePreview 或提交时使用）
   *
   * 策略：
   * - 明文 UTF-8 字节数 ≤ maxRsaPlainBytes → v1（纯 RSA-OAEP）
   * - 明文 UTF-8 字节数 > maxRsaPlainBytes → v2（AES-256-GCM + RSA-OAEP 混合加密）
   *
   * 后端按首字节自动识别版本，无需额外协商。
   */
  encryptSync(plainText: string): string | false {
    if (!plainText || !cachedPublicKey) {
      console.warn('加密失败：明文为空或公钥未加载');
      return false;
    }

    try {
      const publicKey = forge.pki.publicKeyFromPem(cachedPublicKey) as forge.pki.rsa.PublicKey;

      // 计算明文 UTF-8 字节长度
      const plainBytes = forge.util.encodeUtf8(plainText);
      const plainByteLen = plainBytes.length;

      if (plainByteLen <= maxRsaPlainBytes) {
        // v1: 纯 RSA-OAEP 直接加密
        return encryptV1(publicKey, plainText);
      }
      // v2: AES-GCM + RSA-OAEP 混合加密
      return encryptV2(publicKey, plainText);
    } catch (error) {
      console.error('加密过程异常:', error);
      return false;
    }
  },

  /**
   * 3. 强制使用 V1（纯 RSA-OAEP）加密 —— 用于密码
   */
  encryptV1Sync(plainText: string): string | false {
    if (!plainText || !cachedPublicKey) {
      console.warn('V1 加密失败：明文为空或公钥未加载');
      return false;
    }
    try {
      const publicKey = forge.pki.publicKeyFromPem(cachedPublicKey) as forge.pki.rsa.PublicKey;
      return encryptV1(publicKey, plainText);
    } catch (error) {
      console.error('V1 加密过程异常:', error);
      return false;
    }
  },

  /**
   * 4. 强制使用 V2（AES-256-GCM + RSA-OAEP 混合）加密 —— 用于密钥
   */
  encryptV2Sync(plainText: string): string | false {
    if (!plainText || !cachedPublicKey) {
      console.warn('V2 加密失败：明文为空或公钥未加载');
      return false;
    }
    try {
      const publicKey = forge.pki.publicKeyFromPem(cachedPublicKey) as forge.pki.rsa.PublicKey;
      return encryptV2(publicKey, plainText);
    } catch (error) {
      console.error('V2 加密过程异常:', error);
      return false;
    }
  },
};
