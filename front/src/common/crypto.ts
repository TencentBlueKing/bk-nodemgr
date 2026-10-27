import { Message } from 'bkui-vue';
import JSEncrypt from 'jsencrypt';
import { sm2 } from 'sm-crypto';

import { gcm } from '@noble/ciphers/aes.js';
import { randomBytes } from '@noble/ciphers/utils.js';
import { sha256 } from '@noble/hashes/sha2.js';

import { CipherService } from '@/api/modules/cipher';
import { i18n } from '@/modules/i18n';

const VERSION_V1 = 0x01;
const VERSION_V2 = 0x02;
const VERSION_SM2 = 0x03;
const SM2_CIPHER_MODE_C1C3C2 = 1;
const RSA_OAEP_LABEL = 'com.example.crypto.rsa.v1';
const AES_KEY_BYTES = 32;
const AES_NONCE_BYTES = 12;
const AES_TAG_BYTES = 16;

interface RsaPublicKey {
  n: bigint;
  e: bigint;
  keyBytes: number;
}

// 当前启用套件的加密函数（初始化时按 key_type 绑定，业务侧无需感知算法）
type ActiveEncryptor = (plainText: string) => string | false;

let activeEncryptor: ActiveEncryptor | null = null;
let pendingPromise: Promise<boolean> | null = null;

// ---- helpers ----

function uint8ArrayToBase64(bytes: Uint8Array): string {
  let binary = '';
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
  return btoa(binary);
}

function base64ToUint8Array(base64: string): Uint8Array {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}

function bytesToHex(bytes: Uint8Array): string {
  let hex = '';
  for (const b of bytes) hex += b.toString(16).padStart(2, '0');
  return hex;
}

function hexToBytes(hex: string): Uint8Array {
  const out = new Uint8Array(hex.length / 2);
  for (let i = 0; i < out.length; i += 1) out[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16);
  return out;
}

function calcMaxPlainBytes(keyBytes: number): number {
  const hashBytes = 32;
  return keyBytes - 2 * hashBytes - 2;
}

/** BigInt modular exponentiation */
function modPow(base: bigint, exp: bigint, mod: bigint): bigint {
  let result = 1n;
  base %= mod;
  while (exp > 0n) {
    if (exp & 1n) result = (result * base) % mod;
    exp >>= 1n;
    base = (base * base) % mod;
  }
  return result;
}

/** MGF1 mask generation function (RFC 8017) */
function mgf1(seed: Uint8Array, length: number): Uint8Array {
  const hLen = 32;
  const T = new Uint8Array(length);
  const count = Math.ceil(length / hLen);
  for (let i = 0; i < count; i++) {
    const C = new Uint8Array(4);
    C[0] = (i >>> 24) & 0xff;
    C[1] = (i >>> 16) & 0xff;
    C[2] = (i >>> 8) & 0xff;
    C[3] = i & 0xff;
    const input = new Uint8Array(seed.length + 4);
    input.set(seed, 0);
    input.set(C, seed.length);
    const hash = sha256(input);
    const copyLen = Math.min(hLen, length - i * hLen);
    T.set(hash.slice(0, copyLen), i * hLen);
  }
  return T;
}

/** OAEP encode per RFC 8017 §7.1.1 */
function oaepEncode(
  message: Uint8Array,
  seed: Uint8Array,
  label: Uint8Array,
  k: number,
): Uint8Array {
  const hLen = 32;

  // 1-2. lHash = Hash(L), PS = zero padding
  const lHash = sha256(label);
  const psLen = k - message.length - 2 * hLen - 2;
  if (psLen < 0) throw new Error('Message too long');

  // 3. DB = lHash || PS || 0x01 || M
  const DB = new Uint8Array(hLen + psLen + 1 + message.length);
  DB.set(lHash, 0);
  DB[hLen + psLen] = 0x01;
  DB.set(message, hLen + psLen + 1);

  // 4-5. maskedDB = DB XOR MGF1(seed)
  const dbMask = mgf1(seed, k - hLen - 1);
  const maskedDB = new Uint8Array(DB.length);
  for (let i = 0; i < DB.length; i++) maskedDB[i] = DB[i] ^ dbMask[i];

  // 6-7. maskedSeed = seed XOR MGF1(maskedDB)
  const seedMask = mgf1(maskedDB, hLen);
  const maskedSeed = new Uint8Array(seed.length);
  for (let i = 0; i < seed.length; i++) maskedSeed[i] = seed[i] ^ seedMask[i];

  // 8. EM = 0x00 || maskedSeed || maskedDB
  const EM = new Uint8Array(1 + maskedSeed.length + maskedDB.length);
  EM.set(maskedSeed, 1);
  EM.set(maskedDB, 1 + maskedSeed.length);

  return EM;
}

/** BigInt → exactly k bytes (big-endian, left-padded with zero) */
function bigintToBytes(value: bigint, k: number): Uint8Array {
  const hex = value.toString(16);
  const paddedHex = hex.length % 2 ? `0${hex}` : hex;
  const out = new Uint8Array(k);
  const hexPairs = paddedHex.match(/.{1,2}/g) || [];
  const offset = out.length - hexPairs.length;
  for (let i = 0; i < hexPairs.length; i++) {
    out[offset + i] = parseInt(hexPairs[i], 16);
  }
  return out;
}

/** Uint8Array (big-endian) → BigInt */
function bytesToBigint(bytes: Uint8Array): bigint {
  let hex = '';
  for (const b of bytes) hex += b.toString(16).padStart(2, '0');
  return BigInt(`0x${hex || '00'}`);
}

// ---- RSA suite (CLASSIC) ----

let cachedRsaPublicKey: RsaPublicKey | null = null;
let maxRsaPlainBytes = 0;

/** RSA-OAEP encrypt raw bytes → exactly keyBytes bytes ciphertext */
function rsaOaepEncrypt(data: Uint8Array): Uint8Array {
  if (!cachedRsaPublicKey) throw new Error('Public key not initialized');
  const label = new TextEncoder().encode(RSA_OAEP_LABEL);
  const seed = randomBytes(32);
  const padded = oaepEncode(data, seed, label, cachedRsaPublicKey.keyBytes);
  const m = bytesToBigint(padded);
  const c = modPow(m, cachedRsaPublicKey.e, cachedRsaPublicKey.n);
  return bigintToBytes(c, cachedRsaPublicKey.keyBytes);
}

/** V1 加密：RSA-OAEP（同步，公钥须已初始化） */
function encryptV1(plainText: string): string | false {
  if (!cachedRsaPublicKey) return false;

  try {
    const plainBytes = new TextEncoder().encode(plainText);
    const encrypted = rsaOaepEncrypt(plainBytes);

    const out = new Uint8Array(1 + encrypted.length);
    out.set(encrypted, 1);
    out[0] = VERSION_V1;

    return uint8ArrayToBase64(out);
  } catch {
    return false;
  }
}

/** V2 加密：AES-256-GCM + RSA-OAEP 封装密钥（同步，公钥须已初始化） */
function encryptV2(plainText: string): string | false {
  if (!cachedRsaPublicKey) return false;

  try {
    // AES-256 key
    const aesKeyBytes = randomBytes(AES_KEY_BYTES);

    // random nonce
    const nonce = randomBytes(AES_NONCE_BYTES);

    // AES-GCM encrypt
    const labelBytes = new TextEncoder().encode(RSA_OAEP_LABEL);
    const plainBytes = new TextEncoder().encode(plainText);
    const aes = gcm(aesKeyBytes, nonce, labelBytes);
    const aesEncrypted = aes.encrypt(plainBytes);

    // separate ciphertext and tag (noble's GCM appends tag)
    const ciphertext = aesEncrypted.slice(0, aesEncrypted.length - AES_TAG_BYTES);
    const tag = aesEncrypted.slice(aesEncrypted.length - AES_TAG_BYTES);

    // RSA-OAEP wrap AES key
    const wrappedKey = rsaOaepEncrypt(aesKeyBytes);

    // assemble V2 output: version || wrappedKey || nonce || ciphertext || tag
    const out = new Uint8Array(1 + wrappedKey.length + nonce.length + ciphertext.length + tag.length);
    let pos = 0;
    out[pos++] = VERSION_V2;
    out.set(wrappedKey, pos); pos += wrappedKey.length;
    out.set(nonce, pos); pos += nonce.length;
    out.set(ciphertext, pos); pos += ciphertext.length;
    out.set(tag, pos);

    return uint8ArrayToBase64(out);
  } catch {
    return false;
  }
}

/** RSA 套件加密：按明文长度自动选择 V1 / V2 */
function rsaEncrypt(plainText: string): string | false {
  const plainBytes = new TextEncoder().encode(plainText);
  if (plainBytes.length <= maxRsaPlainBytes) return encryptV1(plainText);
  return encryptV2(plainText);
}

/** 初始化 RSA 公钥（解析 PEM 提取 n/e），成功返回 true */
function initRsaPublicKey(pem: string): boolean {
  // PEM → JSEncrypt → extract n, e
  const encrypt = new JSEncrypt();
  encrypt.setPublicKey(pem);
  const rsaKey = encrypt.getKey();
  if (!rsaKey) return false;
  // jsencrypt RSAKey n/e 为 protected，通过 any 绕过类型检查
  const rsaAny = rsaKey as any;
  if (!rsaAny.n) return false;

  const n = BigInt(`0x${rsaAny.n.toString(16)}`);
  const e = BigInt(rsaAny.e);
  const keyBytes = Math.ceil(rsaAny.n.bitLength() / 8);

  cachedRsaPublicKey = { n, e, keyBytes };
  maxRsaPlainBytes = calcMaxPlainBytes(keyBytes);

  return true;
}

// ---- SM2 suite (SHANGMI) ----

// SM2 公钥（sm-crypto 约定：04‖X‖Y 的 hex，含未压缩点前缀）
let cachedSm2PublicKey = '';

/**
 * 解析 PKIX PEM（SubjectPublicKeyInfo）中的 SM2 未压缩公钥点。
 * 结构：SEQUENCE { SEQUENCE { OID, OID }, BIT STRING }，BIT STRING 内容为 04 ‖ X(32) ‖ Y(32)。
 * 返回 04‖X‖Y 的 hex（sm-crypto 要求含未压缩点前缀），解析失败返回 null。
 */
function parseSm2PublicKeyPem(pem: string): string | null {
  const b64 = pem.replace(/-----[^-]*-----/g, '').replace(/\s+/g, '');
  if (!b64) return null;

  let der: Uint8Array;
  try {
    der = base64ToUint8Array(b64);
  } catch {
    return null;
  }

  let pos = 0;
  const readByte = (): number => {
    if (pos >= der.length) throw new Error('truncated');
    const b = der[pos];
    pos += 1;
    return b;
  };
  // DER definite-length 读取（后端生成的公钥不会出现 indefinite form）
  const readLength = (): number => {
    const first = readByte();
    if (first < 0x80) return first;
    const n = first & 0x7f;
    if (n === 0 || n > 4) throw new Error('unsupported length');
    let len = 0;
    for (let i = 0; i < n; i++) len = len * 256 + readByte();
    return len;
  };
  const expectTag = (tag: number): number => {
    if (readByte() !== tag) throw new Error('unexpected tag');
    return readLength();
  };

  try {
    // SubjectPublicKeyInfo: SEQUENCE { AlgorithmIdentifier, BIT STRING }
    const spkiLen = expectTag(0x30);
    if (pos + spkiLen > der.length) throw new Error('truncated');

    const algLen = expectTag(0x30);
    pos += algLen; // 跳过 AlgorithmIdentifier（算法与曲线由后端保证）

    const bitsLen = expectTag(0x03);
    if (readByte() !== 0) throw new Error('unexpected unused bits');
    if (bitsLen - 1 !== 65 || pos + 65 > der.length) throw new Error('bad point length');
    if (der[pos] !== 0x04) throw new Error('not an uncompressed point');

    return bytesToHex(der.subarray(pos, pos + 65));
  } catch {
    return null;
  }
}

/** SM2 套件加密：密文格式 base64(0x03 ‖ 04 ‖ C1x ‖ C1y ‖ C3 ‖ C2)，无长度分档 */
function sm2Encrypt(plainText: string): string | false {
  if (!cachedSm2PublicKey) return false;

  try {
    const plainBytes = new TextEncoder().encode(plainText);
    // sm-crypto 输出 C1C3C2 拼接的 hex（C1 不含 04 前缀），后端按裸 C1C3C2 格式识别
    const cipherHex = sm2.doEncrypt(plainBytes, cachedSm2PublicKey, SM2_CIPHER_MODE_C1C3C2);
    const cipherBytes = hexToBytes(cipherHex);

    const out = new Uint8Array(2 + cipherBytes.length);
    out[0] = VERSION_SM2;
    out[1] = 0x04; // C1 未压缩点前缀
    out.set(cipherBytes, 2);

    return uint8ArrayToBase64(out);
  } catch {
    return false;
  }
}

/** 初始化 SM2 公钥，成功返回 true */
function initSm2PublicKey(pem: string): boolean {
  const pubHex = parseSm2PublicKeyPem(pem);
  if (!pubHex) return false;

  cachedSm2PublicKey = pubHex;

  return true;
}

// ---- public API ----

// 初始化失败原因对应的提示文案：网络/服务端错误，或服务端套件前端不认识
const MSG_INIT_FAILED = 'message.credentialEncrypt.initFailed';
const MSG_UNSUPPORTED_SUITE = 'message.credentialEncrypt.unsupportedSuite';

// 上一次初始化失败的提示文案，供 encrypt 在真正提交时把原因反馈给用户
let initFailureMessage: string | null = null;

function showError(messageKey: string): void {
  Message({ theme: 'error', message: i18n.global.t(messageKey) });
}

/**
 * 拉取后端当前套件公钥并初始化（异步，仅此一步需要网络请求）。
 * 按 key_type 绑定对应算法：RSA4096（CLASSIC）或 SM2（SHANGMI），
 * 算法选择由服务端全局配置决定；未知 key_type 显式失败，不静默回退。
 *
 * 本函数不弹提示：页面挂载时的预热调用不是用户动作，提示由 encrypt 在真正
 * 提交凭据时统一负责，避免在用户尚未提交任何凭据时就弹错。
 */
async function initPublicKey(): Promise<boolean> {
  if (activeEncryptor) return true;
  if (pendingPromise) return pendingPromise;

  pendingPromise = (async () => {
    let res: { key_type?: string; public_key?: string } | null = null;
    try {
      res = await CipherService.GetCurrentPublicKey({});
    } catch {
      initFailureMessage = MSG_INIT_FAILED;
      return false;
    }

    const keyType = res?.key_type ?? '';
    const pem = res?.public_key || '';
    if (!pem) {
      initFailureMessage = MSG_INIT_FAILED;
      return false;
    }

    switch (keyType) {
      case 'RSA4096':
        if (!initRsaPublicKey(pem)) {
          initFailureMessage = MSG_UNSUPPORTED_SUITE;
          return false;
        }
        activeEncryptor = rsaEncrypt;
        break;
      case 'SM2':
        if (!initSm2PublicKey(pem)) {
          initFailureMessage = MSG_UNSUPPORTED_SUITE;
          return false;
        }
        activeEncryptor = sm2Encrypt;
        break;
      default:
        // 未知算法（后端配置了前端不认识的套件），显式失败而非静默回退
        initFailureMessage = MSG_UNSUPPORTED_SUITE;
        return false;
    }

    initFailureMessage = null;
    return true;
  })();

  try {
    return await pendingPromise;
  } finally {
    pendingPromise = null;
  }
}

/**
 * 统一加密入口：按当前套件加密，RSA 自动选择 V1/V2，SM2 单层加密（无长度限制）。
 *
 * 初始化失败或加密失败都会显式提示并返回 false，调用方据此中止提交；
 * 业务侧不应把 false 当作空凭据继续提交。
 */
async function encrypt(plainText: string): Promise<string | false> {
  if (!activeEncryptor && !(await initPublicKey())) {
    showError(initFailureMessage ?? MSG_INIT_FAILED);
    return false;
  }

  // initPublicKey 成功即表示 activeEncryptor 已绑定，这里取本地引用收窄类型
  const encryptor = activeEncryptor;
  if (!encryptor) {
    showError(MSG_INIT_FAILED);
    return false;
  }

  const encrypted = encryptor(plainText);
  if (encrypted === false) {
    showError('message.credentialEncrypt.encryptFailed');
  }

  return encrypted;
}

export {
  initPublicKey,
  encrypt,
};
