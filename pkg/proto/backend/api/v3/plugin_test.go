package v3

import "testing"

func TestPluginInstallReqConvertParamToTypesWithHostBizMappingPreservesCompatibilityMode(t *testing.T) {
	req := &PluginInstallReq{
		EnableCompatibilityMode: true,
		Plugin: []*PluginOperateFullInfo{
			{
				BkHostId:   1001,
				PluginName: "bkmonitorbeat",
				Version:    "1.0.0",
			},
		},
	}

	params := req.ConvertParamToTypesWithHostBizMapping(map[int64]int64{1001: 2001})
	if len(params) != 1 {
		t.Fatalf("expected 1 plugin deployment param, got %d", len(params))
	}

	if !params[0].EnableCompatibilityMode {
		t.Fatalf("expected compatibility mode to be preserved")
	}
}
