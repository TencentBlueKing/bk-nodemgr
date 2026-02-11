package frontsetting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeFrontValueByKind(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		kind     frontValueKind
		input    string
		expected string
	}{
		{name: "url-empty", kind: frontValueKindURL, input: "", expected: ""},
		{name: "url-no-trailing-slash", kind: frontValueKindURL, input: "https://bknode.com/path", expected: "https://bknode.com/path"},
		{name: "url-trailing-slash", kind: frontValueKindURL, input: "https://bknode.com/path/", expected: "https://bknode.com/path"},
		{name: "url-multiple-trailing-slashes", kind: frontValueKindURL, input: "https://bknode.com/path///", expected: "https://bknode.com/path"},
		{name: "url-keep-scheme-only", kind: frontValueKindURL, input: "http://", expected: "http://"},
		{name: "uri-root", kind: frontValueKindURI, input: "/", expected: "/"},
		{name: "uri-trailing-slash", kind: frontValueKindURI, input: "/api/v1/", expected: "/api/v1"},
		{name: "host-direct", kind: frontValueKindHost, input: "example.com/", expected: "example.com"},
		{name: "host-from-url", kind: frontValueKindHost, input: "https://example.com/path/", expected: "example.com"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, normalizeFrontValue(testCase.input, testCase.kind))
		})
	}
}

func TestNewFrontSettingNormalizesURLs(t *testing.T) {
	t.Parallel()

	setting, err := NewFrontSetting(Option{
		BKLoginURL:            "https://login.example.com/",
		BKRequestIDHeaderKEy:  "X-Request-Id",
		BKUserWebURL:          "https://user.example.com/",
		BKDomain:              "https://domain.example.com///",
		BKDocsCenterURL:       "https://docs.example.com/",
		BKAppNavOpenSourceURL: "https://github.com/TencentBlueKing/bk-nodemgr/",
	})
	require.NoError(t, err)

	require.Equal(t, "https://login.example.com", setting.BKLoginURL())
	require.Equal(t, "https://user.example.com", setting.BKUserWebURL())
	require.Equal(t, "domain.example.com", setting.BKDomain())
	require.Equal(t, "https://docs.example.com", setting.BKDocsCenterURL())
	require.Equal(t, "https://github.com/TencentBlueKing/bk-nodemgr", setting.BKAppNavOpenSourceURL())
}
