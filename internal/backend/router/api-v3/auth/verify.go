package auth

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
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

	err := h.verifyItems(rCtx, req.ConvertItemsToCheckItems())
	if err != nil {
		logger.G.Biz(rCtx).WithErr(err).Error("failed to verify auth")

		var permErr auth.PermissionDeniedError
		if errors.As(err, &permErr) {
			return nil, resterrf.ErrWrap(resterrf.PermissionDenied, err)
		}

		return nil, resterrf.ErrWrap(resterrf.BackendOperateFailed, err)
	}

	resp := new(protoBackend.AuthVerifyResp)
	resp.ConvertResultsFromVerify(req.ConvertItemsToVerifyResults())

	return resp.GetData(), nil
}

func (h *handler) verifyItems(
	rCtx restserver.IContext, items []protoBackend.AuthCheckItem,
) error {
	for _, item := range items {
		err := h.verifyItem(rCtx, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *handler) verifyItem(
	rCtx restserver.IContext, item protoBackend.AuthCheckItem,
) error {
	checkErr := h.authorizer.Check(rCtx, item.Action, item.Resources)
	if checkErr == nil {
		return nil
	}

	var permErr auth.PermissionDeniedError
	if errors.As(checkErr, &permErr) {
		return checkErr
	}

	return fmt.Errorf("check action %s permission: %w", item.Action, checkErr)
}
