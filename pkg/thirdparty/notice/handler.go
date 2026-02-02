/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package notice

import (
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	restclient "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/client"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/scheduler"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

const (
	taskNameRegisterApp = "notice_register_app"
	registerAppInterval = scheduler.Daily
	registerAppTimeout  = time.Minute
)

// IHandler the Handler of notice.
type IHandler interface {
	IAnnouncement
}

// IAnnouncement this interface is used to get announcement info.
type IAnnouncement interface {
	// GetCurrentAnnouncements retrieves current active announcements.
	// Platform is obtained from config (appCode), username from context.
	GetCurrentAnnouncements(nCtx contextx.IContext) ([]*types.Announcement, error)
}

// Handler the Handler of notice.
type Handler struct {
	cli       *cli
	scheduler scheduler.Scheduler
}

// Verify that Handler implements IHandler interface.
var _ IHandler = (*Handler)(nil)

// New initialize a new notice Handler.
func New(c *restclient.Capability, conf *Config) (*Handler, error) {
	cli, err := newClient(c, conf)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		cli:       cli,
		scheduler: scheduler.NewScheduler(),
	}

	// Initialize and start scheduler (failures only log warnings)
	h.initScheduler()

	return h, nil
}

// initScheduler initializes the scheduler for periodic app registration.
// Failures are logged as warnings and do not prevent Handler creation.
func (h *Handler) initScheduler() {
	logger.G.Sys().Info("initializing notice app registration scheduler")

	task := scheduler.NewTask(
		taskNameRegisterApp,
		registerAppInterval,
		registerAppTimeout,
		h.registerAppTask,
	)

	if err := h.scheduler.RegisterTask(task); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to register notice app task")
		return
	}

	// Execute immediately on startup
	if err := h.registerAppTask(contextx.Background()); err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to register notice app on startup")
	}

	h.scheduler.Start()
	logger.G.Sys().Info("notice app registration scheduler started")
}

// registerAppTask is the periodic task function for registering the application.
func (h *Handler) registerAppTask(nCtx contextx.IContext) error {
	registration, err := h.registerApplication(nCtx)
	if err != nil {
		logger.G.Sys().WithErr(err).Warn("failed to register notice app")
		return nil // Don't return error, allow retry next time
	}

	logger.G.Sys().With("app-id", registration.ID).Info("registered notice app")

	return nil
}

// registerApplication registers the application with notice center.
// This is an internal method, not exposed in the interface.
func (h *Handler) registerApplication(nCtx contextx.IContext) (*appRegistration, error) {
	if nCtx == nil {
		return nil, errors.New("failed to register application: context is nil")
	}

	resp, err := h.cli.registerApplication(nCtx)
	if err != nil {
		return nil, err
	}

	return &appRegistration{
		ID:   resp.ID,
		Code: resp.Code,
		Name: resp.Name,
	}, nil
}

// GetCurrentAnnouncements retrieves current active announcements.
func (h *Handler) GetCurrentAnnouncements(nCtx contextx.IContext) ([]*types.Announcement, error) {
	if nCtx == nil {
		return nil, errors.New("failed to get current announcements: context is nil")
	}

	// Build params from config and context
	params := &getCurrentAnnouncementsParams{
		Platform: h.cli.config.APIGWUserConfig.GetAppCode(),
	}

	resp, err := h.cli.getCurrentAnnouncements(nCtx, params)
	if err != nil {
		return nil, err
	}

	// Convert response to types.Announcement
	announcements := make([]*types.Announcement, len(resp))
	for idx, ann := range resp {
		startTime, err := parseTime(ann.StartTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start_time for announcement %d: %w", ann.ID, err)
		}

		endTime, err := parseTime(ann.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end_time for announcement %d: %w", ann.ID, err)
		}

		announcements[idx] = &types.Announcement{
			ID:           ann.ID,
			Title:        ann.Content.Title,
			Content:      ann.Content.Content,
			AnnounceType: ann.AnnounceType,
			StartTime:    startTime,
			EndTime:      endTime,
		}
	}

	return announcements, nil
}

// parseTime parses a time string to time.Time.
// Returns error if parsing fails with all supported formats.
func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}

	// Try common time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse time string: %s", timeStr)
}
