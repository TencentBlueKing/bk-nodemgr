package v3

import (
	"testing"
	"time"

	platfmt "github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageReleaseAgentListBriefRespConvertReleasesFromTypes(t *testing.T) {
	resp := &PackageReleaseAgentListBriefResp{}
	resp.ConvertReleasesFromTypes(1, []*types.ReleaseAgent{{
		Release: types.Release{
			Generation: types.Generation(2),
			Platform:   platfmt.Platform{OS: criteria.OSType("linux"), Arch: criteria.CPUArch("amd64")},
			Version:    "2.0.1",
			Enabled:    true,
			AsDefault:  false,
			Name:       "ignored-name",
			FileName:   "ignored-file",
			Labels:     []string{"ignored"},
			MD5:        "ignored-md5",
			UpdatedAt:  time.UnixMilli(1710000000000),
			Operator:   "ignored-operator",
			Type:       types.ReleaseTypeAgent,
		},
		ReleaseAdditionInfoAgent: types.ReleaseAdditionInfoAgent{
			ChangeLogEN: "ignored-en",
			ChangeLogZH: "ignored-zh",
		},
	}})

	require.NotNil(t, resp.Data)
	require.Len(t, resp.Data.Items, 1)
	item := resp.Data.Items[0]
	assert.Equal(t, int64(2), item.GetGeneration())
	assert.Equal(t, "linux", item.GetOsType())
	assert.Equal(t, "amd64", item.GetCpuArch())
	assert.Equal(t, "2.0.1", item.GetVersion())
	assert.True(t, item.GetEnabled())
	assert.False(t, item.GetAsDefault())
}

func TestPackageReleaseProxyListBriefRespConvertReleasesToTypes(t *testing.T) {
	resp := &PackageReleaseProxyListBriefResp{
		Data: &PackageReleaseProxyListBriefResp_Data{
			Total: 1,
			Items: []*ReleaseProxyBrief{{
				Generation:  int64Ptr(3),
				OsType:      stringPtr("linux"),
				CpuArch:     stringPtr("arm64"),
				Version:     stringPtr("3.2.1"),
				Enabled:     boolPtr(true),
				AsDefault:   boolPtr(true),
				ChangeLogEn: stringPtr("en"),
				ChangeLogZh: stringPtr("zh"),
			}},
		},
	}

	total, releases := resp.ConvertReleasesToTypes()
	require.Len(t, releases, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, types.Generation(3), releases[0].Generation)
	assert.Equal(t, criteria.OSType("linux"), releases[0].Platform.OS)
	assert.Equal(t, criteria.CPUArch("arm64"), releases[0].Platform.Arch)
	assert.Equal(t, "3.2.1", releases[0].Version)
	assert.True(t, releases[0].Enabled)
	assert.True(t, releases[0].AsDefault)
	assert.Equal(t, "en", releases[0].ChangeLogEN)
	assert.Equal(t, "zh", releases[0].ChangeLogZH)
}

func TestPackageReleasePluginListBriefRespConvertReleasePluginsFromTypes(t *testing.T) {
	resp := &PackageReleasePluginListBriefResp{}
	resp.ConvertReleasePluginsFromTypes(1, []*types.ReleasePlugin{{
		Release: types.Release{
			Name:       "bk-plugin",
			Generation: types.Generation(5),
			Type:       types.ReleaseTypePlugin,
			Platform:   platfmt.Platform{OS: criteria.OSType("windows"), Arch: criteria.CPUArch("x86_64")},
			Version:    "5.0.0",
			FileName:   "plugin.zip",
			Labels:     []string{"prod"},
			Enabled:    true,
			AsDefault:  false,
			MD5:        "plugin-md5",
			UpdatedAt:  time.UnixMilli(1710000000456),
			Operator:   "tester",
		},
	}})

	require.NotNil(t, resp.Data)
	require.Len(t, resp.Data.Items, 1)
	item := resp.Data.Items[0]
	assert.Equal(t, int64(5), item.GetGeneration())
	assert.Equal(t, "windows", item.GetOsType())
	assert.Equal(t, "x86_64", item.GetCpuArch())
	assert.Equal(t, "5.0.0", item.GetVersion())
	assert.True(t, item.GetEnabled())
	assert.False(t, item.GetAsDefault())
}

func int64Ptr(v int64) *int64    { return &v }
func boolPtr(v bool) *bool       { return &v }
func stringPtr(v string) *string { return &v }
