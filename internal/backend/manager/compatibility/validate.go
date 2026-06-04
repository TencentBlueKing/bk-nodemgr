package compatibility

import (
	"errors"
	"fmt"
)

// Validate checks whether the policy is structurally valid after normalization.
func (p Policy) Validate() error {
	if p.EnabledPlugins == nil {
		return errors.New("enabled_plugins is missing")
	}
	if p.DisabledBiz == nil {
		return errors.New("disabled_biz is missing")
	}

	policy := p.Normalize()

	for _, pluginName := range policy.EnabledPlugins {
		if pluginName == "" {
			return errors.New("enabled_plugins item is empty")
		}
	}

	for _, item := range policy.DisabledBiz {
		if item.TenantID == "" {
			return errors.New("disabled_biz tenant_id is empty")
		}
		if item.BKBizID <= 0 {
			return fmt.Errorf("disabled_biz bk_biz_id is invalid: %d", item.BKBizID)
		}
	}

	return nil
}

// ValidateForWrite strictly validates a policy before persisting it.
func ValidateForWrite(policy Policy) error {
	return policy.Validate()
}
