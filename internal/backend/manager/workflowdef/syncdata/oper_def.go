package syncdata

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/operengine"
)

const (

	// OperDefNameSyncBiz sync all biz from cmdb.
	OperDefNameSyncBiz = "oper_def_sync_biz"
)

// NewOperSyncBizFromCMDB new an operation to sync all biz's host from cmdb.
func NewOperSyncBizFromCMDB(triggerID string) *operengine.Operation {
	defSnapshot := operengine.OperDefSnapshot{
		OperDefName: OperDefNameSyncBiz,
		ActionNames: []string{ActionNameSyncBizFromCMDB},
	}

	operation := operengine.NewOperation("", defSnapshot)

	return operation
}
