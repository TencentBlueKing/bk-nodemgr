package auth

import (
	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

// Authorized queries the authorized resource scope for requested action-resource type pairs.
func (h *handler) Authorized(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.AuthorizedReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query authorized scope, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	scopes, err := h.queryAuthorizedScopes(rCtx, req.GetItems())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to query authorized scope")
		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.AuthorizedResp)
	resp.ConvertResultsFromScopes(req.GetItems(), scopes)

	return resp.GetData(), nil
}

func (h *handler) queryAuthorizedScopes(
	rCtx restserver.IContext, items []*protoBackend.AuthorizedItem,
) ([]auth.AuthorizedScope, error) {
	if len(items) == 0 {
		return nil, nil
	}

	scopes := make([]auth.AuthorizedScope, len(items))
	for i, item := range items {
		action := auth.Action(item.GetAction())
		resourceType := auth.ResourceType(item.GetResourceType())

		scope, err := h.authorizer.ListAuthorizedInstances(rCtx, action, resourceType)
		if err != nil {
			return nil, err
		}

		scopes[i] = scope
	}

	return scopes, nil
}
