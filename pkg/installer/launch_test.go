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

package installer

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestRenderLaunchScriptOutput(t *testing.T) {
	tests := []struct {
		name       string
		workDir    string
		scriptName string
		render     func(string, func() (string, string, error)) (string, error)
		want       string
	}{
		{
			name: "unix", workDir: "/workspace", scriptName: "install.sh", render: RenderUnixLaunchScript,
			want: `mkdir -p /workspace && cd /workspace && echo "installer --token $TOKEN" > install.sh && sh install.sh`,
		},
		{
			name: "windows batch", workDir: `C:\work dir`, scriptName: "install.bat", render: RenderWindowsLaunchScript,
			want: `mkdir "C:\work dir" 2>nul & cd "C:\work dir" && ` +
				`echo installer --token $TOKEN > "install.bat" && cmd /c "install.bat"`,
		},
		{
			name: "windows command", workDir: `C:\work dir`, scriptName: "install.bat", render: RenderWindowsCommand,
			want: `cd C:\work dir && installer --token $TOKEN`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := tt.render(tt.workDir, func() (string, string, error) {
				return tt.scriptName, "installer --token $TOKEN", nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if command != tt.want {
				t.Errorf("command = %q, want %q", command, tt.want)
			}
		})
	}
}

func TestRenderUnixScriptOutput(t *testing.T) {
	command, err := RenderUnixScript(func() (string, string, error) {
		return "install.sh", "installer --token $TOKEN", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `echo "installer --token $TOKEN" > install.sh && sh install.sh`
	if command != want {
		t.Errorf("command = %q, want %q", command, want)
	}
}

func TestUnixWorkspaceCommandOutput(t *testing.T) {
	script, err := RenderUnixScript(func() (string, string, error) {
		return "install.sh", "installer --token $TOKEN", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	command, err := UnixWorkspaceCommand("/workspace", script)
	if err != nil {
		t.Fatal(err)
	}
	want := `mkdir -p /workspace && cd /workspace && echo "installer --token $TOKEN" > install.sh && sh install.sh`
	if command != want {
		t.Errorf("command = %q, want %q", command, want)
	}
}

func TestRenderLaunchScriptPreservesRenderError(t *testing.T) {
	cause := errors.New("invalid installer parameters")
	tests := []struct {
		name   string
		render func(string, func() (string, string, error)) (string, error)
	}{
		{name: "unix", render: RenderUnixLaunchScript},
		{name: "windows batch", render: RenderWindowsLaunchScript},
		{name: "windows command", render: RenderWindowsCommand},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := tt.render("workspace", func() (string, string, error) {
				return "", "", cause
			})
			if !errors.Is(err, cause) {
				t.Errorf("error = %v, want %v", err, cause)
			}
			if command != "" {
				t.Errorf("command = %q, want empty command", command)
			}
		})
	}
}

func TestWindowsSSHLaunchCommandOutput(t *testing.T) {
	command, err := WindowsSSHLaunchCommand(`C:\work's dir`, `C:\work's dir\install.bat`)
	if err != nil {
		t.Fatal(err)
	}
	want := `Start-Process -FilePath 'cmd.exe' -ArgumentList '/c "C:\work''s dir\install.bat"' ` +
		`-WorkingDirectory 'C:\work''s dir'`
	if got := decodeLaunchPowerShell(t, command); got != want {
		t.Errorf("PowerShell script = %q, want %q", got, want)
	}
}

func TestWindowsShellLaunchCommandOutput(t *testing.T) {
	command, err := WindowsShellLaunchCommand(`C:\work dir`, "install.sh")
	if err != nil {
		t.Fatal(err)
	}
	want := `cd "C:/work dir" && sh "install.sh"`
	if command != want {
		t.Errorf("command = %q, want %q", command, want)
	}
}

func TestEnsureWindowsWorkDirCommandOutput(t *testing.T) {
	command, err := EnsureWindowsWorkDirCommand(`C:\work's dir`)
	if err != nil {
		t.Fatal(err)
	}
	want := `New-Item -ItemType Directory -Force -Path 'C:\work''s dir' | Out-Null`
	if got := decodeLaunchPowerShell(t, command); got != want {
		t.Errorf("PowerShell script = %q, want %q", got, want)
	}
}

func TestInstallerTransportCommandOutput(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		render func(string) (string, error)
		want   string
	}{
		{
			name: "cygwin directory", path: "C:/work dir", render: EnsureWindowsShellWorkDirCommand,
			want: `mkdir -p "C:/work dir"`,
		},
		{
			name: "unix executable", path: "/work/installer", render: UnixMakeExecutableCommand,
			want: "chmod +x /work/installer",
		},
		{name: "unix directory", path: "/work/config", render: UnixEnsureDirectoryCommand, want: "mkdir -p /work/config"},
		{name: "wmi directory", path: `C:\work dir`, render: WindowsEnsureDirectoryCommand, want: `mkdir C:\work dir`},
		{name: "wmi batch", path: `C:\work dir\install.bat`, render: WindowsBatchCommand, want: `C:\work dir\install.bat`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := tt.render(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if command != tt.want {
				t.Errorf("command = %q, want %q", command, tt.want)
			}
		})
	}
}

func decodeLaunchPowerShell(t *testing.T, command string) string {
	t.Helper()
	const prefix = "powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -EncodedCommand "
	if !strings.HasPrefix(command, prefix) {
		t.Fatalf("unexpected PowerShell command: %q", command)
	}
	encoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(command, prefix))
	if err != nil {
		t.Fatal(err)
	}
	const (
		bytesPerCodeUnit = 2
		bitsPerByte      = 8
	)
	if len(encoded)%bytesPerCodeUnit != 0 {
		t.Fatal("PowerShell command is not UTF-16LE")
	}
	codeUnits := make([]uint16, len(encoded)/bytesPerCodeUnit)
	for idx := range codeUnits {
		codeUnits[idx] = uint16(encoded[idx*bytesPerCodeUnit]) | uint16(encoded[idx*bytesPerCodeUnit+1])<<bitsPerByte
	}

	return string(utf16.Decode(codeUnits))
}
