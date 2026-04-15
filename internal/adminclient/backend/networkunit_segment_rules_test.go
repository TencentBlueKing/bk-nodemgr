/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	backendadmin "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backendadmin"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBackendAdminHandler struct {
	backendadmin.IHandler
	rules       types.NetworkUnitSegmentRuleConfig
	upsertedCfg types.NetworkUnitSegmentRuleConfig
	err         error
}

func (f *fakeBackendAdminHandler) GetNetworkUnitSegmentRules(contextx.IContext) (types.NetworkUnitSegmentRuleConfig, error) {
	return f.rules, f.err
}

func (f *fakeBackendAdminHandler) UpsertNetworkUnitSegmentRules(
	_ contextx.IContext,
	cfg types.NetworkUnitSegmentRuleConfig,
) error {
	f.upsertedCfg = cfg
	return f.err
}

func TestNewBackendCMDRegistersNetworkUnitSegmentRulesSubcommand(t *testing.T) {
	cmd := NewBackendCMD(func(string) (backendadmin.IHandler, error) {
		return &backendadmin.Handler{}, nil
	})

	require.NotNil(t, cmd)
	assert.Equal(t, "backend", cmd.Use)
	assert.NotNil(t, findSubcommand(cmd, "networkunit_segment_rules"))
}

func findSubcommand(cmd *cobra.Command, use string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Use == use {
			return child
		}
	}

	return nil
}

func TestNetworkUnitSegmentRulesGetCommandPrintsRulesJSON(t *testing.T) {
	handler := &fakeBackendAdminHandler{rules: types.NetworkUnitSegmentRuleConfig{
		"0": {
			Rules: []types.NetworkUnitSegmentRule{{
				CIDRs:         []string{"9.135.144.0/24"},
				NetworkUnitID: 1,
			}},
		},
	}}

	cmd := NewNetworkUnitSegmentRulesCMD(
		func() backendadmin.IHandler { return handler },
		func() (string, string) { return "t", "admin" },
	)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"get"})
	cmd.SetContext(contextx.New(context.Background(), contextx.WithTenantID("t"), contextx.WithLoginName("admin"), contextx.WithBKUsername("admin")))

	err := cmd.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `"0"`)
	assert.Contains(t, out.String(), `"9.135.144.0/24"`)
}

func TestNetworkUnitSegmentRulesCommandRegistersUpsertSubcommand(t *testing.T) {
	cmd := NewNetworkUnitSegmentRulesCMD(
		func() backendadmin.IHandler { return &backendadmin.Handler{} },
		func() (string, string) { return "default", "admin" },
	)

	assert.NotNil(t, findSubcommand(cmd, "upsert"))
}

func TestNetworkUnitSegmentRulesUpsertCommandReadsRulesFile(t *testing.T) {
	handler := &fakeBackendAdminHandler{}
	rulesFile := t.TempDir() + "/rules.json"
	require.NoError(t, os.WriteFile(rulesFile, []byte(`{"0":{"rules":[{"cidrs":["0.0.0.0/0"],"bk_networkunit_id":0}]}}`), 0o600))

	cmd := NewNetworkUnitSegmentRulesCMD(
		func() backendadmin.IHandler { return handler },
		func() (string, string) { return "t", "admin" },
	)
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"upsert", "--rules-file", rulesFile})
	cmd.SetContext(contextx.New(context.Background(), contextx.WithTenantID("t"), contextx.WithLoginName("admin"), contextx.WithBKUsername("admin")))

	err := cmd.Execute()

	require.NoError(t, err)
	require.Contains(t, handler.upsertedCfg, "0")
	assert.Equal(t, int64(0), handler.upsertedCfg["0"].Rules[0].NetworkUnitID)
	assert.Contains(t, out.String(), "Successfully upserted networkunit segment rules")
}
