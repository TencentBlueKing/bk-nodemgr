package auth

import (
	"errors"
	"fmt"

	authx "github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	protoBackend "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/backend/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
)

func (h *handler) Verify(rCtx restserver.IContext) (interface{}, error) {
	req := new(protoBackend.AuthVerifyReq)
	if err := rCtx.BindJSON(req); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth, failed to decode request body")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth, invalid request")

		return nil, resterrf.ErrWrap(resterrf.InvalidParameter, err)
	}

	results, err := h.verifyItems(rCtx, req.GetItems())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth")

		var permErr authx.PermissionDeniedError
		if errors.As(err, &permErr) {
			return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
		}

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.AuthVerifyResp)
	resp.ConvertResultsFromVerify(results)

	return resp.GetData(), nil
}

func (h *handler) verifyItems(
	rCtx restserver.IContext, items []*protoBackend.AuthVerifyItem,
) ([]*protoBackend.AuthVerifyResult, error) {

	results := make([]*protoBackend.AuthVerifyResult, 0, len(items))

	for _, item := range items {
		result, err := h.verifyItem(rCtx, item)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (h *handler) verifyItem(
	rCtx restserver.IContext, item *protoBackend.AuthVerifyItem,
) (*protoBackend.AuthVerifyResult, error) {

	action := authx.Action(item.GetAction())
	resources := protoBackend.ConvertAuthResourcesToInternal(item.GetResources())

	checkErr := h.authorizer.Check(rCtx, action, resources)
	if checkErr == nil {
		return protoBackend.NewAuthVerifyResult(item.GetAction(), true), nil
	}

	var permErr authx.PermissionDeniedError
	if errors.As(checkErr, &permErr) {
		return nil, checkErr
	}

	return nil, fmt.Errorf("check action %s permission: %w", item.GetAction(), checkErr)
}
