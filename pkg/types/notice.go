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

package types

import "time"

// AnnounceType represents the type of announcement.
type AnnounceType string

const (
	// AnnounceTypeEvent represents activity/event notification.
	AnnounceTypeEvent AnnounceType = "event"
	// AnnounceTypeAnnounce represents platform announcement.
	AnnounceTypeAnnounce AnnounceType = "announce"
)

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
	// AnnounceType is the type of announcement.
	AnnounceType AnnounceType
	// StartTime is when the announcement becomes active.
	StartTime time.Time
	// EndTime is when the announcement expires.
	EndTime time.Time
}
