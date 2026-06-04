package compatibility

// Policy defines the compatibility policy schema.
type Policy struct {
	EnabledPlugins []string      `json:"enabled_plugins"`
	DisabledBiz    []DisabledBiz `json:"disabled_biz"`
}

// DisabledBiz defines a business scope disabled by tenant and biz ID.
type DisabledBiz struct {
	TenantID string `json:"tenant_id"`
	BKBizID  int64  `json:"bk_biz_id"`
}

// DefaultPolicy returns the default compatibility policy.
func DefaultPolicy() Policy {
	return Policy{
		EnabledPlugins: []string{"bkmonitorbeat"},
		DisabledBiz:    []DisabledBiz{},
	}
}
