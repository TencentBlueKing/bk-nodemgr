/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package config

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config/kms/envelope"
)

func validCloudSecurityGroupPool() CloudSecurityGroupPool {
	return CloudSecurityGroupPool{
		PoolID:           "pool-a",
		PoolName:         "Pool A",
		Vendor:           "tencent",
		Region:           "ap-guangzhou",
		SecretID:         "secret-id",
		SecretKey:        "secret-key",
		SecurityGroupIDs: []string{"sg-1"},
	}
}

func TestCloudSecurityGroupValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*CloudSecurityGroup)
		wantErr bool
	}{
		{name: "valid", mutate: func(*CloudSecurityGroup) {}},
		{name: "duplicate pool id", mutate: func(conf *CloudSecurityGroup) {
			conf.Pools = append(conf.Pools, validCloudSecurityGroupPool())
		}, wantErr: true},
		{name: "missing secret key", mutate: func(conf *CloudSecurityGroup) {
			conf.Pools[0].SecretKey = ""
		}, wantErr: true},
		{name: "unsupported vendor", mutate: func(conf *CloudSecurityGroup) {
			conf.Pools[0].Vendor = "other"
		}, wantErr: true},
		{name: "blank security group id", mutate: func(conf *CloudSecurityGroup) {
			conf.Pools[0].SecurityGroupIDs = []string{""}
		}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := CloudSecurityGroup{Pools: []CloudSecurityGroupPool{validCloudSecurityGroupPool()}}
			tt.mutate(&conf)
			if err := conf.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloudSecurityGroupValidateEmpty(t *testing.T) {
	conf := CloudSecurityGroup{}
	if err := conf.Validate(); err != nil {
		t.Fatalf("empty pools should be valid: %v", err)
	}
}

func TestApplyCloudSecurityGroupSecretKeys(t *testing.T) {
	tests := []struct {
		name          string
		yamlSecretKey string
		envelopePools map[string]envelope.BackendCloudSecurityGroupPoolSecret
		wantSecretKey string
		wantErr       bool
	}{
		{
			name:          "envelope overrides yaml secret key",
			yamlSecretKey: "secret-key",
			envelopePools: map[string]envelope.BackendCloudSecurityGroupPoolSecret{
				"pool-a": {SecretKey: "kms-secret"},
			},
			wantSecretKey: "kms-secret",
		},
		{
			name:          "envelope fills an empty yaml secret key",
			yamlSecretKey: "",
			envelopePools: map[string]envelope.BackendCloudSecurityGroupPoolSecret{
				"pool-a": {SecretKey: "kms-secret"},
			},
			wantSecretKey: "kms-secret",
		},
		{
			name:          "missing envelope entry keeps yaml secret key",
			yamlSecretKey: "secret-key",
			envelopePools: map[string]envelope.BackendCloudSecurityGroupPoolSecret{},
			wantSecretKey: "secret-key",
		},
		{
			name:          "nil envelope pools keeps yaml secret key",
			yamlSecretKey: "secret-key",
			envelopePools: nil,
			wantSecretKey: "secret-key",
		},
		{
			name:          "empty envelope secret key keeps yaml secret key",
			yamlSecretKey: "secret-key",
			envelopePools: map[string]envelope.BackendCloudSecurityGroupPoolSecret{
				"pool-a": {SecretKey: ""},
			},
			wantSecretKey: "secret-key",
		},
		{
			name:          "other pool entry does not satisfy this pool",
			yamlSecretKey: "",
			envelopePools: map[string]envelope.BackendCloudSecurityGroupPoolSecret{
				"pool-b": {SecretKey: "kms-secret"},
			},
			wantErr: true,
		},
		{
			name:          "missing envelope entry without yaml secret key fails",
			yamlSecretKey: "",
			envelopePools: map[string]envelope.BackendCloudSecurityGroupPoolSecret{},
			wantErr:       true,
		},
		{
			name:          "nil envelope pools without yaml secret key fails",
			yamlSecretKey: "",
			envelopePools: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := validCloudSecurityGroupPool()
			pool.SecretKey = tt.yamlSecretKey
			pools := []CloudSecurityGroupPool{pool}

			err := applyCloudSecurityGroupSecretKeys(pools, tt.envelopePools)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if pools[0].SecretKey != tt.wantSecretKey {
				t.Fatalf("secret key = %q, want %q", pools[0].SecretKey, tt.wantSecretKey)
			}
		})
	}
}
