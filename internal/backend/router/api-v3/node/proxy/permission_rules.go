package proxy

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)


func buildBizIDsFromTypeHosts(hosts []*types.Host) []int64 {
	bizIDMap := make(map[int64]struct{})
	for _, host := range hosts {
		if host == nil || host.Static == nil {
			continue
		}
		bizIDMap[host.Static.BizID] = struct{}{}
	}

	return conv.MapKeyToSlice(bizIDMap)
}

