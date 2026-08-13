package etcddiscover

import (
	"strings"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
)

func TestProviderEtcdRegisterPrevalidatesAllInstances(t *testing.T) {
	provider := NewProviderEtcd(&config.Etcd{})
	validInstance := discover.Instance{
		ID:   "valid-instance",
		Name: "valid instance",
	}
	invalidInstance := discover.Instance{
		Name: "missing id",
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Register() panicked before validating all instances: %v", recovered)
		}
	}()

	if err := provider.Register(discover.ServiceNameBackend, validInstance, invalidInstance); err != nil {
		if !strings.Contains(err.Error(), "invalid instance ID") {
			t.Fatalf("Register() error = %v, want invalid instance ID", err)
		}

		return
	}

	t.Fatal("Register() error is nil, want invalid instance error")
}
