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

package v3

import (
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/conv"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// ConvertAnnouncementFromTypes converts types.Announcement to proto Announcement.
func ConvertAnnouncementFromTypes(announcement *types.Announcement) *Announcement {
	if announcement == nil {
		return nil
	}

	// Convert content list using conv.SliceToSlice
	contentList := conv.SliceToSlice(announcement.ContentList, func(content types.AnnouncementContent) *AnnouncementContent {
		return &AnnouncementContent{
			Content:  content.Content,
			Language: content.Language,
		}
	})

	return &Announcement{
		Id:           announcement.ID,
		Title:        announcement.Title,
		ContentList:  contentList,
		Content:      announcement.Content,
		AnnounceType: string(announcement.AnnounceType),
		StartTime:    announcement.StartTime.Format(time.RFC3339),
		EndTime:      announcement.EndTime.Format(time.RFC3339),
	}
}

// ConvertAnnouncementsFromTypes converts announcements from types to proto.
func (r *GetCurrentAnnouncementsResp) ConvertAnnouncementsFromTypes(announcements []*types.Announcement) {
	items := make([]*Announcement, 0, len(announcements))
	for _, announcement := range announcements {
		items = append(items, ConvertAnnouncementFromTypes(announcement))
	}
	r.Data = items
}
