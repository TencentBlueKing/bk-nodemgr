/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package relayhandler

import (
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
)

func TestServerMessagerConfigStringRedactsSecrets(t *testing.T) {
	conf := ServerMessagerConfig{
		SlotID:     1,
		Token:      "slot-token-value",
		AppCode:    "bk_nodemgr",
		AppSecret:  "app-secret-value",
		GSEBaseURL: "https://gse.example.com",
		TLSConfig: &ssl.TLSConfig{
			CA:       []byte("ca-content"),
			Cert:     []byte("cert-content"),
			Key:      []byte("key-content"),
			EncCert:  []byte("enc-cert-content"),
			EncKey:   []byte("enc-key-content"),
			Password: "tls-password-value",
		},
	}

	// the redacted form must not mutate the original config.
	out := conf.String()

	for _, secret := range []string{
		"slot-token-value", "app-secret-value", "tls-password-value",
		"key-content", "enc-key-content",
	} {
		if strings.Contains(out, secret) {
			t.Errorf("String() leaks secret %q", secret)
		}
	}

	if conf.TLSConfig.Key == nil || conf.TLSConfig.EncKey == nil ||
		conf.TLSConfig.Password != "tls-password-value" {
		t.Error("String() must not mutate the original TLSConfig")
	}
}

func TestServerMessagerConfigStringNilTLS(t *testing.T) {
	conf := ServerMessagerConfig{}

	if out := conf.String(); !strings.Contains(out, "TLSConfig:<nil>") {
		t.Errorf("String() with nil TLSConfig = %q", out)
	}
}
