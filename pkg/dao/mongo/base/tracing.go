/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package base provides MongoDB ORM and tracing utilities.
package base

const (
	// spanEventSlowQuery is the event name for slow query tracing.
	spanEventSlowQuery = "orm.slow_query" // nolint:unused

	// Span attribute keys for ORM operations.
	attrKeyORMCollection  = "orm.collection"   // nolint:unused
	attrKeyORMOperation   = "orm.operation"    // nolint:unused
	attrKeyORMDurationMS  = "orm.duration_ms"  // nolint:unused
	attrKeyORMFilterSize  = "orm.filter_size"  // nolint:unused
	attrKeyORMResultCount = "orm.result_count" // nolint:unused
	attrKeyORMFilter      = "orm.filter"
)
