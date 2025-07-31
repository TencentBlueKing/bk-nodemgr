/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package iegtjj

// nolint: gosec
import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
)

// DevicePasswdDecryptor this is used to decrypt the device password only.
type DevicePasswdDecryptor struct {
	password string
}

// NewDecryptor creates a new DevicePasswdDecryptor instance.
func NewDecryptor(password string) *DevicePasswdDecryptor {
	return &DevicePasswdDecryptor{password: password}
}

// getKeyAndIV generates key and initialization vector from salt.
// nolint: nonamedreturns,varnamelen
func (d *DevicePasswdDecryptor) getKeyAndIV(salt []byte, klen int, ilen int) (key []byte, iv []byte) {
	// Default values if not specified
	if klen == 0 {
		klen = 32
	}
	if ilen == 0 {
		ilen = 16
	}

	password := []byte(d.password)
	maxlen := klen + ilen

	// Initial digest
	hash := md5.New() // nolint: gosec
	hash.Write(append(password, salt...))
	keyiv := hash.Sum(nil)

	tmp := [][]byte{keyiv}

	// Generate enough key material
	for len(bytes.Join(tmp, []byte{})) < maxlen {
		hash := md5.New() // nolint: gosec
		hash.Write(append(append(tmp[len(tmp)-1], password...), salt...))
		tmp = append(tmp, hash.Sum(nil))
	}

	keyivCat := bytes.Join(tmp, []byte{})

	return keyivCat[:klen], keyivCat[klen : klen+ilen]
}

// Decrypt decrypts the given ciphertext
// nolint: mnd,varnamelen
func (d *DevicePasswdDecryptor) Decrypt(ciphertext string) (string, error) {
	// Filter input
	var filtered string
	lines := strings.Split(ciphertext, "\n")
	reEmpty := regexp.MustCompile(`^\s*$`)
	reComment := regexp.MustCompile(`^\s*#`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if reEmpty.MatchString(line) || reComment.MatchString(line) {
			continue
		}
		filtered += line + "\n"
	}

	// Base64 decode
	raw, err := base64.StdEncoding.DecodeString(filtered)
	if err != nil {
		return "", err
	}

	// Verify "Salted__" prefix and extract salt
	if !bytes.HasPrefix(raw, []byte("Salted__")) {
		return "", errors.New("invalid data format")
	}
	salt := raw[8:16]

	// Generate key and IV
	key, iv := d.getKeyAndIV(salt, 32, 16)

	// Extract ciphertext
	ciphertextBytes := raw[16:]

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create CBC decrypter
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertextBytes))
	mode.CryptBlocks(plaintext, ciphertextBytes)

	// Remove PKCS7 padding
	paddingLen := int(plaintext[len(plaintext)-1])
	plaintext = plaintext[:len(plaintext)-paddingLen]

	// Trim trailing newlines
	return string(bytes.TrimRight(plaintext, "\n")), nil
}
