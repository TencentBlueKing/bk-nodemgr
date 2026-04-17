package release

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/application/options"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	protoApplication "github.com/TencentBlueKing/bk-nodemgr/pkg/proto/application/api/v3"
	resterrf "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/errf"
	restserver "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/server"
	backendtp "github.com/TencentBlueKing/bk-nodemgr/pkg/thirdparty/backend"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListReleaseAgentBriefProxySuccess(t *testing.T) {
	h := &handler{backendHandler: &fakeReleaseBackendHandler{
		agentBriefResults: []*types.ReleaseAgent{{
			Release: types.Release{Generation: 2, Version: "2.0.1", Enabled: true},
		}},
		agentBriefTotal: 1,
	}}

	data, err := h.ListReleaseAgentBrief(newMockReleaseRestContext(t, &protoApplication.PackageReleaseAgentListBriefReq{}))
	require.NoError(t, err)
	respData, ok := data.(*protoApplication.PackageReleaseAgentListBriefResp_Data)
	require.True(t, ok)
	require.Len(t, respData.Items, 1)
	assert.Equal(t, "2.0.1", respData.Items[0].GetVersion())
}

func TestListReleaseProxyBriefProxyFailureWrapsThirdpartyError(t *testing.T) {
	h := &handler{backendHandler: &fakeReleaseBackendHandler{proxyBriefErr: errors.New("backend failed")}}

	_, err := h.ListReleaseProxyBrief(newMockReleaseRestContext(t, &protoApplication.PackageReleaseProxyListBriefReq{
		Generation: 2,
		OnlyCount:  true,
	}))
	require.Error(t, err)
	code, _ := resterrf.ErrUnwrap(err)
	assert.Equal(t, resterrf.ThirdpartyRequestFailed, code)
}

func TestLoadRegistersReleaseBriefRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/api/v3")
	Load(rg, &options.Capability{BackendHandler: &fakeReleaseBackendHandler{}})

	for _, path := range []string{
		"/api/v3/release/agent/list/brief",
		"/api/v3/release/proxy/list/brief",
		"/api/v3/release/plugin/list/brief",
	} {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusNotFound, w.Code, path)
	}
}

type fakeReleaseBackendHandler struct {
	backendtp.IHandler
	agentBriefResults  []*types.ReleaseAgent
	agentBriefTotal    int64
	agentBriefErr      error
	proxyBriefResults  []*types.ReleaseProxy
	proxyBriefTotal    int64
	proxyBriefErr      error
	pluginBriefResults []*types.ReleasePlugin
	pluginBriefTotal   int64
	pluginBriefErr     error
}

func (h *fakeReleaseBackendHandler) ListReleaseAgentBrief(
	nCtx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition,
) ([]*types.ReleaseAgent, int64, error) {
	_ = nCtx
	_ = gen
	_ = page
	_ = condition
	return h.agentBriefResults, h.agentBriefTotal, h.agentBriefErr
}

func (h *fakeReleaseBackendHandler) ListReleaseProxyBrief(
	nCtx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition,
) ([]*types.ReleaseProxy, int64, error) {
	_ = nCtx
	_ = gen
	_ = page
	_ = condition
	return h.proxyBriefResults, h.proxyBriefTotal, h.proxyBriefErr
}

func (h *fakeReleaseBackendHandler) ListReleasePluginBrief(
	nCtx contextx.IContext, gen types.Generation, page types.Page, condition *types.ReleaseCondition,
) ([]*types.ReleasePlugin, int64, error) {
	_ = nCtx
	_ = gen
	_ = page
	_ = condition
	return h.pluginBriefResults, h.pluginBriefTotal, h.pluginBriefErr
}

func newMockReleaseRestContext(t *testing.T, body interface{}) restserver.IContext {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	gCtx, _ := gin.CreateTestContext(recorder)
	gCtx.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	gCtx.Request.Header.Set("Content-Type", "application/json")

	return &restserver.Context{
		Context: contextx.New(
			context.Background(),
			contextx.WithTenantID("tenant-test"),
			contextx.WithBKUsername("admin"),
			contextx.WithLoginName("admin"),
			contextx.WithMessageID("request-id"),
		),
		IRequest: restserver.NewRequest(gCtx),
	}
}
