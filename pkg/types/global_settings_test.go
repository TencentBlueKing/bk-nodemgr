package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkUnitSegmentRuleConfigJSONShape(t *testing.T) {
	raw := []byte(`{
		"1001": {
			"rules": [
				{"cidrs": ["10.0.0.0/24", "10.0.1.0/24"], "bk_networkunit_id": 200101},
				{"cidrs": ["0.0.0.0/0"], "bk_networkunit_id": 200199}
			]
		}
	}`)

	var cfg NetworkUnitSegmentRuleConfig
	require.NoError(t, json.Unmarshal(raw, &cfg))
	require.Contains(t, cfg, "1001")
	require.Len(t, cfg["1001"].Rules, 2)

	assert.Equal(t, []string{"10.0.0.0/24", "10.0.1.0/24"}, cfg["1001"].Rules[0].CIDRs)
	assert.Equal(t, int64(200101), cfg["1001"].Rules[0].NetworkUnitID)
	assert.Equal(t, []string{"0.0.0.0/0"}, cfg["1001"].Rules[1].CIDRs)
	assert.Equal(t, int64(200199), cfg["1001"].Rules[1].NetworkUnitID)
}

func TestNetworkUnitSegmentRuleSettingName(t *testing.T) {
	assert.Equal(t, "networkunit_segment_rules", GlobalSettingNameNetworkUnitSegmentRules)
}
