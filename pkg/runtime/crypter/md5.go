/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package crypter ...
package crypter

import (
	"crypto/md5" // nolint:gosec
	"encoding/hex"
)

// MD5Sum calculates the MD5 checksum of the given data and returns it as a hexadecimal string.
// Notice: MD5 is not suitable for cryptographic security purposes.
func MD5Sum(data string) string {
	sum := md5.Sum([]byte(data)) // nolint:gosec

	return hex.EncodeToString(sum[:])
}
