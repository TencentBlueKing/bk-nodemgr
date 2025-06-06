/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodepkg

import (
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/test/basetest"
)

// TestSuite this is the test suite for this package
type TestSuite struct {
	basetest.TestSuit
}

// TestAll test all the test cases in this package
func TestAll(t *testing.T) {
	suit := new(TestSuite)
	suit.AddSetupSuiteFunc(func() {})

	basetest.RunTests(t, suit)
}

// TestFormatPkgName ...
func (suite *TestSuite) TestFormatPkgName() {
	type args struct {
		releaseType types.ReleaseType
		generation  types.Generation
		version     string
		osType      string
		cpuArch     string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid agent package",
			args: args{
				releaseType: types.ReleaseTypeAgent,
				generation:  types.Generation1,
				version:     "1.0.0",
				osType:      "linux",
				cpuArch:     "x86_64",
			},
			want:    "gse_agent-1-1.0.0-linux_x86_64.tgz",
			wantErr: false,
		},
		{
			name: "valid task package",
			args: args{
				releaseType: types.ReleaseType("task"),
				generation:  types.Generation2,
				version:     "2.1.0",
				osType:      "windows",
				cpuArch:     "x86_64",
			},
			want:    "gse_task-2-2.1.0-windows_x86_64.tgz",
			wantErr: true,
		},
		{
			name: "valid proxy package",
			args: args{
				releaseType: types.ReleaseTypeProxy,
				generation:  types.Generation(3),
				version:     "3.5.2",
				osType:      "aix",
				cpuArch:     "powerpc",
			},
			want:    "gse_proxy-3-3.5.2-aix_powerpc.tgz",
			wantErr: true,
		},
		{
			name: "valid arm architecture",
			args: args{
				releaseType: types.ReleaseTypeAgent,
				generation:  types.Generation1,
				version:     "1.0.0",
				osType:      "linux",
				cpuArch:     "aarch64",
			},
			want:    "gse_agent-1-1.0.0-linux_aarch64.tgz",
			wantErr: false,
		},
		{
			name: "invalid node role",
			args: args{
				releaseType: types.ReleaseType("invalid"),
				generation:  types.Generation1,
				version:     "1.0.0",
				osType:      "linux",
				cpuArch:     "x86_64",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid generation",
			args: args{
				releaseType: types.ReleaseTypeAgent,
				generation:  types.Generation(99),
				version:     "1.0.0",
				osType:      "linux",
				cpuArch:     "x86_64",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty version",
			args: args{
				releaseType: types.ReleaseTypeAgent,
				generation:  types.Generation1,
				version:     "",
				osType:      "linux",
				cpuArch:     "x86_64",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty os type",
			args: args{
				releaseType: types.ReleaseTypeAgent,
				generation:  types.Generation1,
				version:     "1.0.0",
				osType:      "",
				cpuArch:     "x86_64",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "empty cpu arch",
			args: args{
				releaseType: types.ReleaseTypeAgent,
				generation:  types.Generation1,
				version:     "1.0.0",
				osType:      "linux",
				cpuArch:     "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "macos package",
			args: args{
				releaseType: types.ReleaseTypeAgent,
				generation:  types.Generation1,
				version:     "1.0.0",
				osType:      "darwin",
				cpuArch:     "x86_64",
			},
			want:    "gse_agent-1-1.0.0-darwin_x86_64.tgz",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			pkgName, err := FormatPkgName(
				tt.args.generation,
				tt.args.releaseType,
				platform.Platform{
					OS:   tt.args.osType,
					Arch: tt.args.cpuArch,
				},
				tt.args.version,
			)

			if tt.wantErr {
				suite.Require().Errorf(err, "expected error, but got nil")

				return
			}

			suite.Require().NoError(err, "unexpected error")

			suite.Require().Equal(tt.want, pkgName, "unexpected package name")
			suite.T().Logf("pkgName: %s", pkgName)
		})
	}
}
