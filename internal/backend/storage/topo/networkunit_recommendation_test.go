package topo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	gsdao "github.com/TencentBlueKing/bk-nodemgr/pkg/dao/mongo/globalsettings"
	pkgglobalsettings "github.com/TencentBlueKing/bk-nodemgr/pkg/globalsettings"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

type fakeGlobalSettingsHandler struct {
	value string
	err   error
}

func (f *fakeGlobalSettingsHandler) Get(_ contextx.IContext, _ string) (*types.GlobalSettings, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &types.GlobalSettings{SettingName: pkgglobalsettings.NetworkUnitSegmentRules, Value: f.value}, nil
}

func (f *fakeGlobalSettingsHandler) Count(contextx.IContext, ...gsdao.OptFn) (int64, error) {
	panic("not implemented")
}

func (f *fakeGlobalSettingsHandler) Exist(contextx.IContext, string) (bool, error) {
	panic("not implemented")
}

func (f *fakeGlobalSettingsHandler) List(contextx.IContext, types.Page, ...gsdao.OptFn) ([]*types.GlobalSettings, int64, error) {
	panic("not implemented")
}

func (f *fakeGlobalSettingsHandler) ListWithoutCount(contextx.IContext, types.Page, ...gsdao.OptFn) ([]*types.GlobalSettings, error) {
	panic("not implemented")
}

func (f *fakeGlobalSettingsHandler) Upsert(contextx.IContext, ...*types.GlobalSettings) error {
	panic("not implemented")
}

func (f *fakeGlobalSettingsHandler) Delete(contextx.IContext, ...string) error {
	panic("not implemented")
}

func TestParseNetworkUnitSegmentRuleConfig(t *testing.T) {
	t.Run("parse valid config", func(t *testing.T) {
		cfg, err := parseNetworkUnitSegmentRuleConfig(`{"1001":{"rules":[{"cidrs":["10.0.0.0/24"],"bk_networkunit_id":200101}]}}`)
		require.NoError(t, err)
		require.Contains(t, cfg, "1001")
		require.Len(t, cfg["1001"].Rules, 1)
		assert.Equal(t, []string{"10.0.0.0/24"}, cfg["1001"].Rules[0].CIDRs)
		assert.Equal(t, int64(200101), cfg["1001"].Rules[0].NetworkUnitID)
	})

	t.Run("reject invalid json", func(t *testing.T) {
		_, err := parseNetworkUnitSegmentRuleConfig("{")
		require.Error(t, err)
	})
}

func TestLoadNetworkUnitSegmentRuleConfig(t *testing.T) {
	s := &Storage{daoGlobalSettings: &fakeGlobalSettingsHandler{value: `{"1001":{"rules":[{"cidrs":["10.0.0.0/24"],"bk_networkunit_id":200101}]}}`}}
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("t"))

	cfg, err := s.loadNetworkUnitSegmentRuleConfig(nCtx)
	require.NoError(t, err)
	require.Contains(t, cfg, "1001")
	require.Len(t, cfg["1001"].Rules, 1)
	assert.Equal(t, int64(200101), cfg["1001"].Rules[0].NetworkUnitID)
}

func TestRecommendNetworkUnitsBySegmentKeepsOrderAndUsesFirstMatch(t *testing.T) {
	items := []*types.NetworkUnitSegmentRecommendationItem{
		{NetworkAreaID: 1001, IP: "10.0.0.8"},
		{NetworkAreaID: 1001, IP: "10.0.0.8"},
		{NetworkAreaID: 1001, IP: "11.0.2.8"},
	}

	rules := types.NetworkUnitSegmentRuleConfig{
		"1001": {
			Rules: []types.NetworkUnitSegmentRule{
				{CIDRs: []string{"10.0.0.0/24"}, NetworkUnitID: 200101},
				{CIDRs: []string{"10.0.0.0/16"}, NetworkUnitID: 200102},
				{CIDRs: []string{"0.0.0.0/0"}, NetworkUnitID: 200199},
			},
		},
	}

	results, err := recommendNetworkUnitsBySegment(rules, items, map[int64]int64{
		200101: 1001,
		200102: 1001,
		200199: 1001,
	})
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.Equal(t, int64(200101), results[0].NetworkUnitID)
	assert.Equal(t, int64(200101), results[1].NetworkUnitID)
	assert.Equal(t, int64(200199), results[2].NetworkUnitID)
	assert.Equal(t, "10.0.0.8", results[0].IP)
	assert.Equal(t, "10.0.0.8", results[1].IP)
	assert.Equal(t, "11.0.2.8", results[2].IP)
}

func TestRecommendNetworkUnitsBySegmentReturnsMinusOneForInvalidOrMissingMatches(t *testing.T) {
	items := []*types.NetworkUnitSegmentRecommendationItem{
		{NetworkAreaID: 1001, IP: "bad-ip"},
		{NetworkAreaID: 2002, IP: "10.0.0.1"},
		{NetworkAreaID: 1001, IP: "10.0.1.8"},
	}

	rules := types.NetworkUnitSegmentRuleConfig{
		"1001": {
			Rules: []types.NetworkUnitSegmentRule{
				{CIDRs: []string{"10.0.0.0/24"}, NetworkUnitID: 200101},
			},
		},
	}

	results, err := recommendNetworkUnitsBySegment(rules, items, map[int64]int64{200101: 1001})
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.Equal(t, int64(-1), results[0].NetworkUnitID)
	assert.Equal(t, int64(-1), results[1].NetworkUnitID)
	assert.Equal(t, int64(-1), results[2].NetworkUnitID)
	assert.NotEmpty(t, results[0].Message)
	assert.NotEmpty(t, results[1].Message)
	assert.NotEmpty(t, results[2].Message)
}

func TestRecommendNetworkUnitsBySegmentMatchesAnyCIDRInOneRule(t *testing.T) {
	items := []*types.NetworkUnitSegmentRecommendationItem{{NetworkAreaID: 1001, IP: "10.0.0.8"}}
	rules := types.NetworkUnitSegmentRuleConfig{
		"1001": {
			Rules: []types.NetworkUnitSegmentRule{
				{CIDRs: []string{"192.168.0.0/16", "10.0.0.0/24"}, NetworkUnitID: 200101},
			},
		},
	}

	results, err := recommendNetworkUnitsBySegment(rules, items, map[int64]int64{200101: 1001})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int64(200101), results[0].NetworkUnitID)
	assert.Equal(t, "matched", results[0].Message)
}

func TestRecommendNetworkUnitsBySegmentReturnsMinusOneForCrossAreaUnit(t *testing.T) {
	items := []*types.NetworkUnitSegmentRecommendationItem{{NetworkAreaID: 1001, IP: "10.0.0.8"}}
	rules := types.NetworkUnitSegmentRuleConfig{
		"1001": {
			Rules: []types.NetworkUnitSegmentRule{
				{CIDRs: []string{"10.0.0.0/24"}, NetworkUnitID: 200201},
			},
		},
	}

	results, err := recommendNetworkUnitsBySegment(rules, items, map[int64]int64{200201: 2002})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int64(-1), results[0].NetworkUnitID)
	assert.Equal(t, "invalid recommended network unit", results[0].Message)
}

func TestRecommendNetworkUnitByNetworkSegmentReturnsSafeResultsWhenRuleConfigInvalid(t *testing.T) {
	s := &Storage{daoGlobalSettings: &fakeGlobalSettingsHandler{value: "{"}}
	nCtx := contextx.New(context.Background(), contextx.WithTenantID("t"))
	items := []*types.NetworkUnitSegmentRecommendationItem{{NetworkAreaID: 1001, IP: "10.0.0.1"}}

	results, err := s.recommendNetworkUnitByNetworkSegment(nCtx, items...)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int64(-1), results[0].NetworkUnitID)
	assert.Equal(t, int64(1001), results[0].NetworkAreaID)
	assert.Equal(t, "10.0.0.1", results[0].IP)
	assert.Contains(t, results[0].Message, "invalid")
}
