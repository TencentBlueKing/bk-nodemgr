package syncdata

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

// OperDefNameSyncHost sync specific biz's host from cmdb.
const OperDefNameSyncHost = "oper_def_sync_host"

const (

	// OperDefNameSyncBiz sync all biz from cmdb.
	OperDefNameSyncBiz = "oper_def_sync_biz"

	// OperDefNameSyncNetworkArea sync networkarea from cmdb.
	OperDefNameSyncNetworkArea = "oper_def_sync_networkarea"
)

// NewOperSyncBizFromCMDB new an operation to sync all biz's host from cmdb.
func NewOperSyncBizFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncBiz,
		ActionNames: []string{ActionNameSyncBizFromCMDB},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}

// NewOperSyncHostFromCMDB new an operation to sync all biz's host from cmdb.
func NewOperSyncHostFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncHost,
		ActionNames: []string{ActionNameSyncHostFromCMDB},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}

// NewOperSyncNetworkAreaFromCMDB new an operation to sync all networkareas from cmdb.
func NewOperSyncNetworkAreaFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncNetworkArea,
		ActionNames: []string{ActionNameSyncNetworkAreaFromCMDB},
	}

	operation := operengine.NewOperation(triggerID, defSnapshot)

	return operation
}
