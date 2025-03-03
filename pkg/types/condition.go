/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

// Condition defines the generic condition settings.
type Condition struct {
}

// ConditionType define the condition type.
type ConditionType string

const (
	// ConditionTypeExactInclude means this condition should be matched including given values in exact mode.
	ConditionTypeExactInclude ConditionType = "exact_include"

	// ConditionTypeExactExclude means this condition should be matched excluding given values in exact mode.
	ConditionTypeExactExclude ConditionType = "exact_exclude"

	// ConditionTypeFuzzyInclude means this condition should be matched including given values in a fuzzy mode.
	ConditionTypeFuzzyInclude ConditionType = "fuzzy_include"

	// ConditionTypeFuzzyExclude means this condition should be matched excluding given values in a fuzzy mode.
	ConditionTypeFuzzyExclude ConditionType = "fuzzy_exclude"
)
