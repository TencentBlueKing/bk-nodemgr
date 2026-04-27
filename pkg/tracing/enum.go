/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package tracing

import "go.opentelemetry.io/otel/attribute"

const (
	// ServiceCategoryKey is the key for service category.
	ServiceCategoryKey = attribute.Key("service.category")
)

// ServiceCategory is the category of service.
type ServiceCategory = string

const (
	// ServiceCategoryHTTP is the category for http.
	ServiceCategoryHTTP ServiceCategory = "http"

	// ServiceCategoryDB is the category for db.
	ServiceCategoryDB ServiceCategory = "db"

	// ServiceCategoryCache is the category for cache.
	ServiceCategoryCache ServiceCategory = "cache"

	// ServiceCategoryAsyncBackend is the category for async backend.
	ServiceCategoryAsyncBackend ServiceCategory = "async_backend"
)
