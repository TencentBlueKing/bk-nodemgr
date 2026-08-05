declare module '@noble/ciphers/aes.js' {
  interface GCMCipher {
    encrypt(plaintext: Uint8Array): Uint8Array;
    decrypt(ciphertext: Uint8Array): Uint8Array;
  }
  export function gcm(key: Uint8Array, nonce: Uint8Array, AAD?: Uint8Array): GCMCipher;
}

declare module '@noble/ciphers/utils.js' {
  export function randomBytes(bytesLength?: number): Uint8Array;
}

declare module '@noble/hashes/sha2.js' {
  export const sha256: {
    (data: Uint8Array): Uint8Array;
    outputLen: number;
    blockLen: number;
    create(): { update(data: Uint8Array): void; digest(): Uint8Array };
  };
}
