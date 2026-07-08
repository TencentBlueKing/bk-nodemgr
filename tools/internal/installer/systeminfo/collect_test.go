//go:build !windows

package systeminfo

import "testing"

func TestParseOSReleasePrettyName(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "quoted pretty name",
			content: `NAME="BlueKing Linux"
PRETTY_NAME="BlueKing Linux 3.0"`,
			want: "BlueKing Linux 3.0",
		},
		{
			name:    "unquoted pretty name",
			content: `PRETTY_NAME=BlueKing Linux`,
			want:    "BlueKing Linux",
		},
		{
			name:    "missing pretty name",
			content: `NAME="BlueKing Linux"`,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOSReleasePrettyName(tt.content)
			if got != tt.want {
				t.Fatalf("parseOSReleasePrettyName(%q) = %q, want %q", tt.content, got, tt.want)
			}
		})
	}
}

func TestUnquoteOSReleaseValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "double quoted value",
			value: `"BlueKing Linux 3.0"`,
			want:  "BlueKing Linux 3.0",
		},
		{
			name:  "quoted value trims inner spaces",
			value: `"  BlueKing Linux 3.0  "`,
			want:  "BlueKing Linux 3.0",
		},
		{
			name:  "unquoted value is preserved after outer trim",
			value: " BlueKing Linux 3.0 ",
			want:  "BlueKing Linux 3.0",
		},
		{
			name:  "empty value",
			value: " ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unquoteOSReleaseValue(tt.value)
			if got != tt.want {
				t.Fatalf("unquoteOSReleaseValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
