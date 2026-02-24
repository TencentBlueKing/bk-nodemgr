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

## CONVENTIONS

* No logging in this package.
* Stateless per-call behavior; safe for concurrent use.
* Keep version bytes / labels next to the algorithm implementation.
* Errors are wrapped with `fmt.Errorf("...: %w", err)` and returned to callers.

## ANTI-PATTERNS

* Do not add key storage/retrieval, configuration, or business logic here.
* Do not change version bytes (`RSAVersion`, `AESVersion`) or `RSADefaultLabel` without coordinating with all callers/clients.
* Do not use `MD5Sum` for security-sensitive purposes.
