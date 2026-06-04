package compatibility

import "strings"

// DecideCompatibilityMode returns whether compatibility mode should be enabled for a plugin.
func DecideCompatibilityMode(policy Policy, tenantID string, bkBizID int64, pluginName string) bool {
	normalized := policy.Normalize()
	pluginName = strings.TrimSpace(pluginName)
	tenantID = strings.TrimSpace(tenantID)

	if isDisabledBiz(normalized.DisabledBiz, tenantID, bkBizID) {
		return false
	}

	for _, enabledPlugin := range normalized.EnabledPlugins {
		if enabledPlugin == pluginName {
			return true
		}
	}

	return false
}

func isDisabledBiz(disabledBiz []DisabledBiz, tenantID string, bkBizID int64) bool {
	for _, item := range disabledBiz {
		if item.TenantID == tenantID && item.BKBizID == bkBizID {
			return true
		}
	}

	return false
}
