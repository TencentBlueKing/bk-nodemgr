/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package auth_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/auth"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

// fakeAuthorizer is a test double for auth.IAuthorizer.
// deniedBizIDs controls which biz IDs are denied.
// applyURL is returned in the PermissionDeniedError.
type fakeAuthorizer struct {
	deniedBizIDs map[string]bool
	applyURL     string
}

func (f *fakeAuthorizer) Check(_ contextx.IContext, action auth.Action, resources []auth.Resource) error {
	for _, r := range resources {
		if f.deniedBizIDs[r.ID] {
			rts := make([]auth.RelatedResourceType, 0, len(resources))
			for _, res := range resources {
				rts = append(rts, auth.RelatedResourceType{
					SystemID: res.SystemID,
					Type:     string(res.Type),
					TypeName: auth.ResourceTypeDisplayName(res.Type),
				})
			}
			return auth.PermissionDeniedError{
				ApplyURL: f.applyURL,
				Actions: []auth.ActionInfo{{
					ID:                   string(action),
					Name:                 auth.ActionDisplayName(action),
					RelatedResourceTypes: rts,
				}},
			}
		}
	}
	return nil
}

// simulateBatchAuth replicates the batch auth pattern used in install/uninstall/upgrade handlers.
type checkOnlyAuthorizer interface {
	Check(ctx contextx.IContext, action auth.Action, resources []auth.Resource) error
}

func simulateBatchAuth(authorizer checkOnlyAuthorizer, bizIDs []int64) error {
	type ctxStub struct{}
	_ = ctxStub{}

	deniedResources := make([]auth.Resource, 0)
	for _, bizID := range bizIDs {
		authErr := authorizer.Check(nil, auth.ActionAgentOperate,
			[]auth.Resource{{SystemID: auth.SystemIDCMDB, Type: auth.ResourceTypeBiz, ID: fmt.Sprintf("%d", bizID)}})
		if authErr != nil {
			deniedResources = append(deniedResources, auth.Resource{
				SystemID: auth.SystemIDCMDB, Type: auth.ResourceTypeBiz, ID: fmt.Sprintf("%d", bizID),
			})
		}
	}
	if len(deniedResources) > 0 {
		aggErr := authorizer.Check(nil, auth.ActionAgentOperate, deniedResources)
		if aggErr == nil {
			aggErr = auth.PermissionDeniedError{Actions: []auth.ActionInfo{{
				ID:   string(auth.ActionAgentOperate),
				Name: auth.ActionDisplayName(auth.ActionAgentOperate),
			}}}
		}
		return aggErr
	}
	return nil
}

func TestBatchBizIDAuth_AllAllowed(t *testing.T) {
	authorizer := &fakeAuthorizer{deniedBizIDs: map[string]bool{}}
	err := simulateBatchAuth(authorizer, []int64{1, 2, 3})
	if err != nil {
		t.Errorf("expected no error when all bizIDs allowed, got: %v", err)
	}
}

func TestBatchBizIDAuth_PartialDenied_AggregatesAllDenied(t *testing.T) {
	authorizer := &fakeAuthorizer{
		deniedBizIDs: map[string]bool{"2": true, "3": true},
		applyURL:     "https://iam.example.com/apply",
	}
	err := simulateBatchAuth(authorizer, []int64{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for partial denied bizIDs, got nil")
	}

	var permErr auth.PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected PermissionDeniedError, got %T: %v", err, err)
	}

	if len(permErr.Actions) == 0 {
		t.Fatal("expected at least one ActionInfo in PermissionDeniedError")
	}

	// Verify aggregated resources contain both denied bizIDs
	rts := permErr.Actions[0].RelatedResourceTypes
	rtIDs := make(map[string]bool)
	for _, rt := range rts {
		rtIDs[rt.Type] = true
	}

	if len(rts) < 2 {
		t.Errorf("expected at least 2 related resource types (for bizID 2 and 3), got %d", len(rts))
	}
}

func TestBatchBizIDAuth_AllDenied(t *testing.T) {
	authorizer := &fakeAuthorizer{
		deniedBizIDs: map[string]bool{"1": true, "2": true, "3": true},
		applyURL:     "https://iam.example.com/apply",
	}
	err := simulateBatchAuth(authorizer, []int64{1, 2, 3})
	if err == nil {
		t.Fatal("expected error when all bizIDs denied, got nil")
	}

	var permErr auth.PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected PermissionDeniedError, got %T: %v", err, err)
	}

	if permErr.ApplyURL != "https://iam.example.com/apply" {
		t.Errorf("expected apply_url %q, got %q", "https://iam.example.com/apply", permErr.ApplyURL)
	}
}

func TestBatchBizIDAuth_SingleDenied(t *testing.T) {
	authorizer := &fakeAuthorizer{
		deniedBizIDs: map[string]bool{"2": true},
		applyURL:     "https://iam.example.com/apply",
	}
	err := simulateBatchAuth(authorizer, []int64{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for single denied bizID, got nil")
	}

	var permErr auth.PermissionDeniedError
	if !errors.As(err, &permErr) {
		t.Fatalf("expected PermissionDeniedError, got %T: %v", err, err)
	}
}

func TestNewNoOpAuthorizer_ExposesCheckMethod(t *testing.T) {
	checkAuthorizer, ok := any(auth.NewNoOpAuthorizer()).(interface {
		Check(contextx.IContext, auth.Action, []auth.Resource) error
	})
	if !ok {
		t.Fatal("expected NewNoOpAuthorizer to expose Check method")
	}

	err := checkAuthorizer.Check(nil, auth.ActionAgentOperate, []auth.Resource{{
		SystemID: auth.SystemIDCMDB,
		Type:     auth.ResourceTypeBiz,
		ID:       "1",
	}})
	if err != nil {
		t.Fatalf("expected nil error from no-op check, got: %v", err)
	}
}

func TestNewNoOpAuthorizer_ExposesListAuthorizedInstancesMethod(t *testing.T) {
	scopeAuthorizer, ok := any(auth.NewNoOpAuthorizer()).(interface {
		ListAuthorizedInstances(contextx.IContext, auth.Action, auth.ResourceType) (auth.AuthorizedScope, error)
	})
	if !ok {
		t.Fatal("expected NewNoOpAuthorizer to expose ListAuthorizedInstances method")
	}

	scope, err := scopeAuthorizer.ListAuthorizedInstances(nil, auth.ActionAgentView, auth.ResourceTypeBiz)
	if err != nil {
		t.Fatalf("expected nil error from no-op list authorized instances, got: %v", err)
	}
	if !scope.IsAny {
		t.Fatal("expected IsAny=true from no-op authorized scope")
	}
	if len(scope.Resources) != 0 {
		t.Fatalf("expected empty Resources from no-op authorized scope, got %+v", scope.Resources)
	}
}
