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

// Package envelope defines the decrypted KMS envelope structures.
package envelope

// GlobalEnvelope defines kms global envelope.
type GlobalEnvelope struct {
	BKAppIDSecret struct {
		Default BKAppIDSecret `json:"default"`
	} `json:"bkapp_id_secret"`

	Etcd struct {
		Default Etcd `json:"default"`
	} `json:"etcd"`

	Redis struct {
		Default Redis `json:"default"`
	} `json:"redis"`

	MongoDB struct {
		Default MongoDB `json:"default"`
	} `json:"mongodb"`

	JwtServerConfig JwtServerConfig `json:"jwt_server_config"`
}

// BKAppIDSecret defines app code and app secret in kms.
type BKAppIDSecret struct {
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
}

// Etcd defines etcd config in kms.
type Etcd struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	TLSCABase64   string `json:"tls_ca_base64"`
	TLSCertBase64 string `json:"tls_cert_base64"`
	TLSKeyBase64  string `json:"tls_key_base64"`
	TLSPassword   string `json:"tls_password"`
}

// Redis defines redis config in kms.
type Redis struct {
	Password string `json:"password"`
}

// MongoDB defines mongodb config in kms.
type MongoDB struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JwtServerConfig defines jwt server config in kms.
type JwtServerConfig struct {
	ApplicationInfo  JwtServerConfigItem `json:"application_info"`
	ApplicationAdmin JwtServerConfigItem `json:"application_admin"`
	ApplicationBasic JwtServerConfigItem `json:"application_basic"`
	BackendInfo      JwtServerConfigItem `json:"backend_info"`
	BackendAdmin     JwtServerConfigItem `json:"backend_admin"`
	BackendBasic     JwtServerConfigItem `json:"backend_basic"`
	BackendCallback  JwtServerConfigItem `json:"backend_callback"`
	BackendProxy     JwtServerConfigItem `json:"backend_proxy"`
	FileInfo         JwtServerConfigItem `json:"file_info"`
	FileAdmin        JwtServerConfigItem `json:"file_admin"`
	FileBasic        JwtServerConfigItem `json:"file_basic"`
	FileDownload     JwtServerConfigItem `json:"file_download"`
	FileExport       JwtServerConfigItem `json:"file_export"`
}

// JwtServerConfigItem defines jwt server config item in kms.
type JwtServerConfigItem struct {
	SymmetricKey       string `json:"symmetric_key"`
	PublicKeyPEMBase64 string `json:"public_key_pem_base64"`
}
