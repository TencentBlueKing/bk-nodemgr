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

import "time"

// AnnouncementContent represents announcement content in a specific language.
type AnnouncementContent struct {
	// Content is the announcement content text.
	Content string
	// Language is the language code (e.g., "zh-cn", "en").
	Language string
}

// Announcement represents a platform announcement from notice center.
type Announcement struct {
	// ID is the unique identifier of the announcement.
	ID int64
	// Title is the announcement title.
	Title string
	// ContentList contains announcement content in multiple languages.
	ContentList []AnnouncementContent
	// Content is the default announcement content (for backward compatibility).
	Content string
	// AnnounceType is the type of announcement (e.g., "maintenance", "update").
	AnnounceType string
	// StartTime is when the announcement becomes active.
	StartTime time.Time
	// EndTime is when the announcement expires.
	EndTime time.Time
}
