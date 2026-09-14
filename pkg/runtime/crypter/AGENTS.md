# CRYPTER KNOWLEDGE BASE

## OVERVIEW

`pkg/runtime/crypter` provides portable cryptographic primitives behind a small interface:

* `Crypter` interface: `Encrypt([]byte) ([]byte, error)` / `Decrypt([]byte) ([]byte, error)`
* Implementations: RSA-OAEP, AES-CBC, plus a non-security MD5 helper.

This package does **not** manage keys, storage, rotation, permissions, or any business rules. Callers are responsible for key retrieval and policy decisions.

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Crypter interface | `crypter.go` |
| RSA key gen / PEM parse / Encrypt-Decrypt | `rsa.go` |
| RSA base64 helper | `rsa.go` → `DecryptRSABase64Ciphertext` |
| AES-CBC crypter | `aes.go` → `NewAESCrypter` |
| SM4-GCM crypter (pure Go) | `sm4.go` → `NewSM4Crypter` |
| SM2 crypter / keygen / PEM parse | `sm2.go` → `NewSM2CrypterFromPrivateKey`, `GenerateSM2KeyPairPEM` |
| MD5 helper (non-security) | `md5.go` → `MD5Sum` |

## RSA (RSA-OAEP)

### Constants / format

* Label: `RSADefaultLabel = "com.example.crypto.rsa.v1"`
* Ciphertext layout (binary): `RSAVersion(1 byte) + raw RSA ciphertext`
* Supported key sizes: 2048 / 3072 / 4096 (`RSAKeySize`)

### Constructors / helpers

* `NewRSACrypterFromPrivateKey(pemBytes []byte, opts ...RSAOption) (Crypter, error)`
  * Enables both encrypt/decrypt (public key derived from private key).
  * Private key PEM parsing supports **PKCS#1** and **PKCS#8** formats.
* `GenerateRSAKeyPairPEM(bits RSAKeySize) (privPEM []byte, pubPEM []byte, err error)`
  * Private key output: PKCS#1 (`"RSA PRIVATE KEY"`).
  * Public key output: PKIX (`"PUBLIC KEY"`).
* `GetRSAPublicKeyPEMByPrivateKeyPEM(privPEM []byte) ([]byte, error)` extracts PKIX public key PEM.

### Base64 helper

* `DecryptRSABase64Ciphertext(cry Crypter, base64Ciphertext string) (string, error)`
  * Expects: `base64.StdEncoding` of `RSAVersion + raw RSA ciphertext`.

## AES (AES-CBC)

### Constants / format

* Version byte: `AESVersion = 1`
* Salt default: `AESDefaultSalt = "com.example.crypto.v1"`
* Ciphertext layout (binary): `AESVersion(1 byte) + IV(16 bytes) + CBC ciphertext`

### Construction

* `NewAESCrypter(key []byte, opts ...Option) (Crypter, error)`
  * Derives a fixed-length 32-byte AES key via SHA-256 over `(key || salt)`.
  * Uses AES-CBC with PKCS7 padding.

### Important behavior

* IV is **deterministically** generated from `(plaintext, derivedKey)`:
  * `IV = sha256(plaintext || derivedKey)[:16]`
  * This means encrypting the same plaintext with the same key produces the same ciphertext.

## SM4 (SM4-GCM, pure Go)

### Constants / format

* Version byte: `SM4Version = 2` (shares the symmetric ciphertext space with `AESVersion = 1`)
* Salt default: `SM4DefaultSalt = "com.example.crypto.sm4.v1"`
* Ciphertext layout (binary): `SM4Version(1 byte) + Nonce(12 bytes) + GCM ciphertext incl. tag(16 bytes)`

### Construction

* `NewSM4Crypter(key []byte, opts ...SM4Option) (Crypter, error)`
  * Backed by emmansun/gmsm (`sm4.NewCipher` + stdlib `cipher.NewGCM`), **pure Go**, no CGO.
  * Derives a fixed-length 16-byte key via `sha256(key || salt)[:16]`.
  * Random 12-byte nonce per encryption.

### Important behavior

* SM4-GCM is authenticated: a wrong key or tampered ciphertext fails `Decrypt` with an error.
* Same plaintext encrypts to different ciphertexts (random nonce); no consumer
  relies on deterministic ciphertext (host credits are upserted by creditID).
* Runtime suite selection: `pkg/config/backend.go` `CryptoType` (`CLASSIC` | `SHANGMI`),
  wired in `internal/backend/service/service.go` → `newSymmetricCrypter`; only the
  configured suite is constructed, the other suite's ciphertexts are rejected by
  their version byte (no cross-suite decrypt dispatch).
* Backend selection rationale: the BlueKing crypto SDK family standardizes the
  interface/config convention, not the backend (Java uses TencentKonaSMSuite,
  Python uses tongsuopy); the Go side picked pure-Go gmsm accordingly.

## SM2 (SM2 public-key encryption, pure Go)

### Constants / format

* Version byte: `SM2Version = 3` (shares the frontend credential ciphertext
  space with RSA v1/v2; only the configured suite is accepted on decrypt)
* Ciphertext layout: `SM2Version(1) + SM2 ciphertext (ASN.1 C1C3C2 DER)`
* Keys: PKCS#8 PEM private / PKIX PEM public, `smx509` based parsing

### Construction / helpers

* `NewSM2CrypterFromPrivateKey(pemBytes)` — both encrypt and decrypt.
* `GenerateSM2KeyPairPEM()` / `GetSM2PublicKeyPEMByPrivateKeyPEM(privPEM)`.
* `DecryptBase64Ciphertext(cry, base64)` — generic base64 helper (the
  RSA-specific one delegates to it).

### Important behavior

* Decrypt auto-detects ASN.1 DER and raw C1C3C2 (04-prefix) encodings.
* No plaintext length limit; wrong key / tampered ciphertext fails with an
  error (SM3 digest check).
* Keypair management: cipher storage keeps RSA4096 and SM2 records under the
  same "DEFAULT" name (unique key: name+keyType); only the keypair of the
  globally configured suite is used at runtime (`config.EncryptCryptoType.CredentialKeyType`).
* Public key distribution: the unified `POST /api/v3/cipher/get_public_key`
  returns the current suite's key + key_type.
* Frontend contract: `docs/developer/credential-encryption-contract.md`.

## CONVENTIONS

* No logging in this package.
* Stateless per-call behavior; safe for concurrent use.
* Keep version bytes / labels next to the algorithm implementation.
* Errors are wrapped with `fmt.Errorf("...: %w", err)` and returned to callers.

## ANTI-PATTERNS

* Do not add key storage/retrieval, configuration, or business logic here.
* Do not change version bytes (`RSAVersion`, `AESVersion`, `SM4Version`) or `RSADefaultLabel` without coordinating with all callers/clients.
* Do not use `MD5Sum` for security-sensitive purposes.
* Do not reintroduce CGO-backed implementations without weighing the build-chain
  cost (cross-compilation, CI toolchains, runtime image portability).
