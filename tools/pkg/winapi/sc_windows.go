//go:build windows

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

package winapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/retrier"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// ServiceInfo 存储Windows服务信息
type ServiceInfo struct {
	Name        string
	DisplayName string
	Status      WinSvcStatus
	StartType   string
	Config      mgr.Config
}

type WinSvcStatus string

const (
	// WinSvcStatusStopped window service status stopped.
	WinSvcStatusStopped WinSvcStatus = "WinSvcStatusStopped"

	// WinSvcStatusStartPending window service status start pending.
	WinSvcStatusStartPending WinSvcStatus = "WinSvcStatusStartPending"

	// WinSvcStatusStopPending window service status stop pending.
	WinSvcStatusStopPending WinSvcStatus = "WinSvcStatusStopPending"

	// WinSvcStatusRunning window service status running.
	WinSvcStatusRunning WinSvcStatus = "WinSvcStatusRunning"

	// WinSvcStatusContinuePending window service status continue pending.
	WinSvcStatusContinuePending WinSvcStatus = "WinSvcStatusContinuePending"

	// WinSvcStatusPausePending window service status pause pending.
	WinSvcStatusPausePending WinSvcStatus = "WinSvcStatusPausePending"

	// WinSvcStatusPaused window service status paused.
	WinSvcStatusPaused WinSvcStatus = "WinSvcStatusPaused"

	// WinSvcStatusUnknown window service status unknown.
	WinSvcStatusUnknown WinSvcStatus = "WinSvcStatusUnknown"

	// WinSvcStatusNotInstalled window service status not installed.
	// this is not a windows service status, but is an actual status.
	WinSvcStatusNotInstalled WinSvcStatus = "WinSvcStatusNotInstalled"
)

func ConvStateToWinSvcStatus(status svc.State) WinSvcStatus {
	switch status {
	case windows.SERVICE_STOPPED:
		return WinSvcStatusStopped
	case windows.SERVICE_START_PENDING:
		return WinSvcStatusStartPending
	case windows.SERVICE_STOP_PENDING:
		return WinSvcStatusStopPending
	case windows.SERVICE_RUNNING:
		return WinSvcStatusRunning
	case windows.SERVICE_CONTINUE_PENDING:
		return WinSvcStatusContinuePending
	case windows.SERVICE_PAUSE_PENDING:
		return WinSvcStatusPausePending
	case windows.SERVICE_PAUSED:
		return WinSvcStatusPaused
	default:
		return WinSvcStatusUnknown
	}
}

// getStartType 将启动类型转换为可读字符串
func getStartType(startType uint32) string {
	switch startType {
	case windows.SERVICE_AUTO_START:
		return "AUTO_START"
	case windows.SERVICE_BOOT_START:
		return "BOOT_START"
	case windows.SERVICE_DEMAND_START:
		return "DEMAND_START"
	case windows.SERVICE_DISABLED:
		return "DISABLED"
	case windows.SERVICE_SYSTEM_START:
		return "SYSTEM_START"
	default:
		return "UNKNOWN"
	}
}

// GetServiceStatus get service status.
func GetServiceStatus(svcName string) (WinSvcStatus, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(svcName)
	if err != nil && !errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return "", fmt.Errorf("failed to open service, svcName(%s): %w", svcName, err)
	}

	if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		return WinSvcStatusNotInstalled, nil
	}

	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return "", fmt.Errorf("failed to get status for service, svcName(%s): %v", svcName, err)
	}

	return ConvStateToWinSvcStatus(status.State), nil
}

// GetService get service info.
func GetService(svcName string) (*ServiceInfo, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(svcName)
	if err != nil {
		return nil, fmt.Errorf("failed to open service, svcName(%s): %w", svcName, err)
	}
	defer s.Close()

	config, err := s.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get config for service, svcName(%s): %v", svcName, err)
	}

	status, err := s.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to get status for service, svcName(%s): %v", svcName, err)

	}

	serviceInfo := &ServiceInfo{
		Name:        svcName,
		DisplayName: config.DisplayName,
		Status:      ConvStateToWinSvcStatus(status.State),
		StartType:   getStartType(config.StartType),
		Config:      config,
	}

	return serviceInfo, nil
}

// StopService stop service.
func StopService(ctx context.Context, svcName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(svcName)
	if err != nil {
		return fmt.Errorf("failed to open service, svcName(%s): %w", svcName, err)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service, svcName(%s): %w", svcName, err)
	}

	r := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())
	err = r.Do(ctx, func(attempt int) error {
		status, err := s.Query()
		if err != nil {
			return fmt.Errorf("failed to get status for service, svcName(%s): %v", svcName, err)
		}

		if status.State != svc.Stopped {
			return fmt.Errorf("service not stopped, svcName(%s), status: %d", svcName, status.State)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to stop service, svcName(%s): %w", svcName, err)
	}

	return nil
}
