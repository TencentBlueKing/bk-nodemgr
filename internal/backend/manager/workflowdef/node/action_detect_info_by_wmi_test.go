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

package node

import (
	"context"
	"errors"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
)

type fakeWindowsWMICommandRunner struct {
	outputs []fakeWindowsWMIOutput
	calls   []string
}

type fakeWindowsWMIOutput struct {
	stdout string
	err    error
}

func (r *fakeWindowsWMICommandRunner) RunCommand(ctx context.Context, cmd string) (string, string, error) {
	r.calls = append(r.calls, cmd)
	if len(r.outputs) == 0 {
		return "", "", errors.New("unexpected command")
	}
	out := r.outputs[0]
	r.outputs = r.outputs[1:]
	return out.stdout, "", out.err
}

func TestDetectWindowsWMIInfo(t *testing.T) {
	tests := []struct {
		name      string
		outputs   []fakeWindowsWMIOutput
		want      windowsWMIDetectResult
		wantCalls []string
		wantErr   bool
	}{
		{
			name: "windows amd64",
			outputs: []fakeWindowsWMIOutput{
				{stdout: "Microsoft Windows [Version 10.0.19045]\n"},
				{stdout: "AMD64\n"},
			},
			want: windowsWMIDetectResult{
				osType:  criteria.OSWindows,
				cpuArch: criteria.CPUArchAmd64,
			},
			wantCalls: []string{"ver", "echo %PROCESSOR_ARCHITECTURE%"},
		},
		{
			name: "ver command failure",
			outputs: []fakeWindowsWMIOutput{
				{err: errors.New("ver failed")},
			},
			wantErr:   true,
			wantCalls: []string{"ver"},
		},
		{
			name: "arch command failure",
			outputs: []fakeWindowsWMIOutput{
				{stdout: "Microsoft Windows [Version 10.0.19045]\n"},
				{err: errors.New("arch failed")},
			},
			wantErr:   true,
			wantCalls: []string{"ver", "echo %PROCESSOR_ARCHITECTURE%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &fakeWindowsWMICommandRunner{outputs: tt.outputs}
			got, err := detectWindowsWMIInfo(context.Background(), runner)
			if (err != nil) != tt.wantErr {
				t.Fatalf("detectWindowsWMIInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("detectWindowsWMIInfo() = %+v, want %+v", got, tt.want)
			}
			if len(runner.calls) != len(tt.wantCalls) {
				t.Fatalf("calls = %+v, want %+v", runner.calls, tt.wantCalls)
			}
			for i := range tt.wantCalls {
				if runner.calls[i] != tt.wantCalls[i] {
					t.Fatalf("calls = %+v, want %+v", runner.calls, tt.wantCalls)
				}
			}
		})
	}
}
