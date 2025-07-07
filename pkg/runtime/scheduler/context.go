/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package scheduler ...
package scheduler

const (
	// Yearly same as '0 0 0 1 1 *'.
	Yearly = "@yearly"

	// Annually same as '0 0 0 1 1 *'.
	Annually = "@annually"

	// Monthly same as '0 0 0 1 * *'.
	Monthly = "@monthly"

	// Weekly same as '0 0 0 * * 0'.
	Weekly = "@weekly"

	// Daily same as '0 0 0 * * *'.
	Daily = "@daily"

	// Midnight same as '0 0 0 * * *'.
	Midnight = "@midnight"

	// Hourly same as '0 0 * * * *'.
	Hourly = "@hourly"

	// Every can be followed by any valid Go time.Duration string.
	Every = "@every "
)
