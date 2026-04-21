/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package provider

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
)

// IHandler provides the iam resource handlers.
type IHandler interface {
	IDispatcher
	IAttributeEnricher
	IResolver
}

// IResolver provides reusable policy-based instance listing over registered providers.
type IResolver interface {
	// ListInstancesByExpression evaluates IAM policy expression (ExprCell format) against provider instances.
	// Used by auth layer fallback when discrete policy parsing fails.
	// Directly accepts expression.ExprCell to avoid unnecessary map serialization round-trip.
	ListInstancesByExpression(
		ctx contextx.IContext,
		resourceType string,
		expr *expression.ExprCell,
		page types.Page,
	) (*ListInstanceData, error)
}

// IDispatcher defines the interface for IAM callback routing.
type IDispatcher interface {
	RegisterProvider(_type string, provider IProvider)
	GetProvider(_type string) (IProvider, bool)
	DispatchMethod(
		ctx contextx.IContext,
		resourceType string,
		method RequestMethod,
		filterMap map[string]interface{},
		page types.Page,
	) (interface{}, error)
}

// IAttributeEnricher provides resource attribute enrichment for authorization.
//
// This interface enables automatic injection of IAM-specific attributes
// (like _bk_iam_path_) into authorization requests before sending to IAM.
//
// Design rationale:
//   - Decouples attribute fetching from authorization logic
//   - Leverages existing provider.FetchInstanceInfo implementations
//   - Supports batch fetching for performance
//
// Integration with iam-go-sdk:
//   - Attributes are set into expression.ObjectSet during policy evaluation
//   - expression.ExprCell.Eval() uses these attributes for policy matching
//   - Special handling for _bk_iam_path_ with starts_with operator (see SDK)
//
// Usage flow:
//  1. auth.Check() calls enrichResourceAttributes() before IAM request
//  2. enrichResourceAttributes() groups resources by (systemID, type)
//  3. For each group, calls FetchResourceAttributes() to get attributes
//  4. Merges fetched attributes into types.AuthResource.Attributes
//  5. Converts to types.IAMResource and sends to IAM SDK
type IAttributeEnricher interface {
	// FetchResourceAttributes fetches attributes for a batch of resources.
	//
	// Parameters:
	//   - ctx: Request context
	//   - resourceType: Resource type identifier (e.g., "networkunit", "package")
	//   - resourceIDs: List of resource IDs to fetch attributes for
	//
	// Returns:
	//   - map[resourceID]map[attrKey]attrValue: Attributes by resource ID
	//   - error: Any error during fetching
	//
	// Behavior:
	//   - If a resource is not found, it should not be in the returned map
	//   - If a resource has no attributes, it should not be in the returned map
	//   - Empty map (not nil) should be returned if no attributes found
	FetchResourceAttributes(
		ctx contextx.IContext,
		resourceType string,
		resourceIDs []string,
	) (map[string]map[string]interface{}, error)
}
