import JSEncrypt from 'jsencrypt';
import { gcm } from '@noble/ciphers/aes.js';
import { randomBytes } from '@noble/ciphers/utils.js';
import { sha256 } from '@noble/hashes/sha2.js';

import { CipherService } from '@/api/modules/cipher';

const VERSION_V1 = 0x01;
const VERSION_V2 = 0x02;
const RSA_OAEP_LABEL = 'com.example.crypto.rsa.v1';
const AES_KEY_BYTES = 32;
const AES_NONCE_BYTES = 12;
const AES_TAG_BYTES = 16;

interface RsaPublicKey {
  n: bigint;
  e: bigint;
  keyBytes: number;
}

let cachedPublicKey: RsaPublicKey | null = null;
let maxRsaPlainBytes = 0;
let pendingPromise: Promise<boolean> | null = null;

// ---- helpers ----

function uint8ArrayToBase64(bytes: Uint8Array): string {
  let binary = '';
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
  return btoa(binary);
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
  const paddedHex = hex.length % 2 ? '0' + hex : hex;
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
  return BigInt('0x' + (hex || '00'));
}

// ---- RSA operation ----

/** RSA-OAEP encrypt raw bytes → exactly keyBytes bytes ciphertext */
function rsaOaepEncrypt(data: Uint8Array): Uint8Array {
  if (!cachedPublicKey) throw new Error('Public key not initialized');
  const label = new TextEncoder().encode(RSA_OAEP_LABEL);
  const seed = randomBytes(32);
  const padded = oaepEncode(data, seed, label, cachedPublicKey.keyBytes);
  const m = bytesToBigint(padded);
  const c = modPow(m, cachedPublicKey.e, cachedPublicKey.n);
  return bigintToBytes(c, cachedPublicKey.keyBytes);
}

// ---- public API ----

/** 拉取后端公钥并解析（异步，仅此一步需要网络请求） */
async function initPublicKey(): Promise<boolean> {
  if (cachedPublicKey) return true;
  if (pendingPromise) return pendingPromise;

  pendingPromise = (async () => {
    try {
      const res = await CipherService.GetRSAPublicKey({}).catch(() => ({ public_key: '' }));
      const pem = res.public_key || '';
      if (!pem) {
        pendingPromise = null;
        return false;
      }

      // PEM → JSEncrypt → extract n, e
      const encrypt = new JSEncrypt();
      encrypt.setPublicKey(pem);
      const rsaKey = encrypt.getKey();
      if (!rsaKey) {
        pendingPromise = null;
        return false;
      }
      // jsencrypt RSAKey n/e 为 protected，通过 any 绕过类型检查
      const rsaAny = rsaKey as any;
      if (!rsaAny.n) {
        pendingPromise = null;
        return false;
      }

      const n = BigInt('0x' + rsaAny.n.toString(16));
      const e = BigInt(rsaAny.e);
      const keyBytes = Math.ceil(rsaAny.n.bitLength() / 8);

      cachedPublicKey = { n, e, keyBytes };
      maxRsaPlainBytes = calcMaxPlainBytes(keyBytes);
      pendingPromise = null;
      return true;
    } catch {
      pendingPromise = null;
      return false;
    }
  })();

  return pendingPromise;
}

/** V1 加密：RSA-OAEP（同步，公钥须已初始化） */
function encryptV1(plainText: string): string | false {
  if (!cachedPublicKey) return false;

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
  if (!cachedPublicKey) return false;

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

/** 自动选择 V1 / V2 */
async function encrypt(plainText: string): Promise<string | false> {
  if (!cachedPublicKey) return false;
  const plainBytes = new TextEncoder().encode(plainText);
  if (plainBytes.length <= maxRsaPlainBytes) return encryptV1(plainText);
  return encryptV2(plainText);
}

export {
  initPublicKey,
  encrypt,
  encryptV1,
  encryptV2,
};
