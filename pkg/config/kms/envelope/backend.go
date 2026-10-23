/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package envelope

// BackendEnvelope defines kms backend envelope.
type BackendEnvelope struct {
	Encryption  BackendEncryption  `json:"encryption"`
	GSE         BackendGSE         `json:"gse"`
	CreditVault BackendCreditVault `json:"credit_vault"`
}

// BackendEncryption defines kms backend encryption.
type BackendEncryption struct {
	EncryptKey string `json:"encrypt_key"`
}

// BackendGSE defines kms backend gse.
type BackendGSE struct {
	PluginSlotID    int    `json:"plugin_slot_id"`
	PluginSlotToken string `json:"plugin_slot_token"`
}

// BackendCreditVault defines kms backend credit vault.
type BackendCreditVault struct {
	IEGTJJ BackendIEGTJJ `json:"iegtjj"`
}

// BackendIEGTJJ defines kms backend iegtjj.
type BackendIEGTJJ struct {
	Key       string `json:"key"`
	SecretKey string `json:"secret_key"`
}
