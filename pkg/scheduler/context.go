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

	// Every1s defines the schedule interval of every 1 second.
	Every1s = Every + "1s"

	// Every5s defines the schedule interval of every 5 seconds.
	Every5s = Every + "5s"

	// Every10s defines the schedule interval of every 10 seconds.
	Every10s = Every + "10s"

	// Every30s defines the schedule interval of every 30 seconds.
	Every30s = Every + "30s"

	// Every1m defines the schedule interval of every 1 minute.
	Every1m = Every + "1m"

	// Every5m defines the schedule interval of every 5 minutes.
	Every5m = Every + "5m"

	// Every10m defines the schedule interval of every 10 minutes.
	Every10m = Every + "10m"

	// Every30m defines the schedule interval of every 30 minutes.
	Every30m = Every + "30m"

	// Every1h defines the schedule interval of every 1 hour.
	Every1h = Every + "1h"

	// Every10h defines the schedule interval of every 10 hours.
	Every10h = Every + "10h"
)
