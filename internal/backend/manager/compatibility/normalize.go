package compatibility

import "strings"

// Normalize returns a normalized copy of the policy.
func (p Policy) Normalize() Policy {
	normalized := Policy{
		EnabledPlugins: normalizeStrings(p.EnabledPlugins),
		DisabledBiz:    make([]DisabledBiz, 0, len(p.DisabledBiz)),
	}

	seen := make(map[DisabledBiz]struct{}, len(p.DisabledBiz))
	for _, item := range p.DisabledBiz {
		normalizedItem := DisabledBiz{
			TenantID: strings.TrimSpace(item.TenantID),
			BKBizID:  item.BKBizID,
		}
		if _, ok := seen[normalizedItem]; ok {
			continue
		}
		seen[normalizedItem] = struct{}{}
		normalized.DisabledBiz = append(normalized.DisabledBiz, normalizedItem)
	}

	if normalized.EnabledPlugins == nil {
		normalized.EnabledPlugins = []string{}
	}
	if normalized.DisabledBiz == nil {
		normalized.DisabledBiz = []DisabledBiz{}
	}

	return normalized
}

func normalizeStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}
