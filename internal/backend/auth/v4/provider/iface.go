/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package provider

import (
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// IHandler composes registered resource queries and runtime enrichment.
type IHandler interface {
	IDispatcher
	IQueryHandler
	IAttributeEnricher
	RegisterProvider(resourceType string, provider IProvider)
	GetProvider(resourceType string) (IProvider, bool)
}

// IDispatcher dispatches callback methods without exposing transport types.
type IDispatcher interface {
	DispatchMethod(
		ctx contextx.IContext, resourceType string, method RequestMethod,
		filterMap map[string]interface{}, page types.Page, requires []string,
	) (interface{}, error)
}

// IInstanceLister enumerates candidates for authorized parent-scope expansion.
type IInstanceLister interface {
	ListInstance(ctx contextx.IContext, resourceType string, req *Request[ListInstanceFilter]) (*ListInstanceData, error)
}

// IQueryHandler exposes typed callback queries without transport dispatch.
type IQueryHandler interface {
	IInstanceLister
	FetchInstanceInfo(ctx contextx.IContext, resourceType string, req *Request[FetchInstanceFilter]) (*FetchInstanceInfoData, error)
}

// IAttributeEnricher fetches attributes in bounded batches for runtime authorization.
type IAttributeEnricher interface {
	FetchResourceAttributes(ctx contextx.IContext, resourceType string, resourceIDs []string) (map[string]map[string]interface{}, error)
}
