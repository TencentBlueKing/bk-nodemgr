package networkarea

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableName_UsesTenantSuffix(t *testing.T) {
	require.Equal(t, "networkarea_tenant-a", TableName("tenant-a"))
}
