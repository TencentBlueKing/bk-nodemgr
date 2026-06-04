package compatibility

import (
	"fmt"
	"testing"
)

func TestCompatibilityPolicyBehavior(t *testing.T) {
	t.Run("DefaultAllowlist", func(t *testing.T) {
		policy := DefaultPolicy()
		if err := ValidateForWrite(policy); err != nil {
			t.Fatalf("default policy should be valid: %v", err)
		}

		allow := DecideCompatibilityMode(policy, "tenant-a", 1, "bkmonitorbeat")
		deny := DecideCompatibilityMode(policy, "tenant-a", 1, "other-plugin")
		if !allow || deny {
			t.Fatalf("unexpected decision result: allow=%v deny=%v", allow, deny)
		}

		t.Logf("default_allowlist allow=%v deny=%v policy=%s", allow, deny, policyString(policy))
	})

	t.Run("DisabledBizOverride", func(t *testing.T) {
		policy := Policy{
			EnabledPlugins: []string{"bkmonitorbeat", "bkdata"},
			DisabledBiz:    []DisabledBiz{{TenantID: "tenant-a", BKBizID: 2}},
		}
		if err := ValidateForWrite(policy); err != nil {
			t.Fatalf("override policy should be valid: %v", err)
		}

		blocked := DecideCompatibilityMode(policy, "tenant-a", 2, "bkmonitorbeat")
		allowed := DecideCompatibilityMode(policy, "tenant-a", 3, "bkmonitorbeat")
		otherTenant := DecideCompatibilityMode(policy, "tenant-b", 2, "bkmonitorbeat")
		if blocked || !allowed || !otherTenant {
			t.Fatalf("unexpected override decision result: blocked=%v allowed=%v otherTenant=%v", blocked, allowed, otherTenant)
		}

		t.Logf("disabled_biz_override blocked=%v allowed=%v other_tenant=%v", blocked, allowed, otherTenant)
	})

	t.Run("ValidEmptyPolicy", func(t *testing.T) {
		policy := ParseStoredPolicy(`{"enabled_plugins":[],"disabled_biz":[]}`, nil)
		if err := ValidateForWrite(policy); err != nil {
			t.Fatalf("empty policy should be valid: %v", err)
		}
		if len(policy.EnabledPlugins) != 0 || len(policy.DisabledBiz) != 0 {
			t.Fatalf("empty policy should stay empty, got %#v", policy)
		}

		t.Logf("valid_empty_policy policy=%s", policyString(policy))
	})

	t.Run("MissingFieldsFallback", func(t *testing.T) {
		for name, raw := range map[string]string{
			"missing_enabled_plugins": `{"disabled_biz":[]}`,
			"missing_disabled_biz":    `{"enabled_plugins":[]}`,
			"empty_object":            `{}`,
		} {
			t.Run(name, func(t *testing.T) {
				policy := ParseStoredPolicy(raw, nil)
				if policyString(policy) != policyString(DefaultPolicy()) {
					t.Fatalf("missing field payload should fall back to default, got %#v", policy)
				}
			})
		}
	})

	t.Run("NullFieldsFallback", func(t *testing.T) {
		for name, raw := range map[string]string{
			"null_enabled_plugins": `{"enabled_plugins":null,"disabled_biz":[]}`,
			"null_disabled_biz":    `{"enabled_plugins":[],"disabled_biz":null}`,
		} {
			t.Run(name, func(t *testing.T) {
				policy := ParseStoredPolicy(raw, nil)
				if policyString(policy) != policyString(DefaultPolicy()) {
					t.Fatalf("null field payload should fall back to default, got %#v", policy)
				}
			})
		}
	})

	t.Run("MalformedFallback", func(t *testing.T) {
		policy := ParseStoredPolicy(`{"enabled_plugins":[`, nil)
		if policyString(policy) != policyString(DefaultPolicy()) {
			t.Fatalf("malformed policy should fall back to default, got %#v", policy)
		}

		t.Logf("malformed_fallback policy=%s", policyString(policy))
	})

	t.Run("InvalidSchemaFallback", func(t *testing.T) {
		policy := ParseStoredPolicy(`{"enabled_plugins":["bkmonitorbeat"],"disabled_biz":[{"tenant_id":"","bk_biz_id":1}]}`, nil)
		if policyString(policy) != policyString(DefaultPolicy()) {
			t.Fatalf("invalid schema should fall back to default, got %#v", policy)
		}

		t.Logf("invalid_schema_fallback policy=%s", policyString(policy))
	})
}

func TestValidateForWriteRejectsInvalidBusinessID(t *testing.T) {
	for _, id := range []int64{0, -1} {
		t.Run(fmt.Sprintf("bk_biz_id_%d", id), func(t *testing.T) {
			policy := Policy{
				EnabledPlugins: []string{"bkmonitorbeat"},
				DisabledBiz:    []DisabledBiz{{TenantID: "tenant-a", BKBizID: id}},
			}
			if err := ValidateForWrite(policy); err == nil {
				t.Fatalf("expected validation error for bk_biz_id=%d", id)
			}
		})
	}
}

func TestValidateForWriteRejectsMissingOrNullArrays(t *testing.T) {
	for name, policy := range map[string]Policy{
		"missing_all":        {},
		"missing_enabled":    {DisabledBiz: []DisabledBiz{}},
		"missing_disabled":   {EnabledPlugins: []string{}},
		"valid_empty_arrays": {EnabledPlugins: []string{}, DisabledBiz: []DisabledBiz{}},
	} {
		t.Run(name, func(t *testing.T) {
			err := ValidateForWrite(policy)
			if name == "valid_empty_arrays" {
				if err != nil {
					t.Fatalf("explicit empty arrays should be valid: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected validation error for %s", name)
			}
		})
	}
}

func policyString(policy Policy) string {
	return fmt.Sprintf("{enabled_plugins:%v disabled_biz:%v}", policy.EnabledPlugins, policy.DisabledBiz)
}
