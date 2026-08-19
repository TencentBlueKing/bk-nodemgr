/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package profiling starts process-level continuous profiling.
package profiling

import (
	"fmt"
	"maps"
	"strings"
	"sync"

	pyroscope "github.com/grafana/pyroscope-go"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
)

var defaultProfileTypes = []pyroscope.ProfileType{
	pyroscope.ProfileCPU,
	pyroscope.ProfileAllocObjects,
	pyroscope.ProfileAllocSpace,
	pyroscope.ProfileInuseObjects,
	pyroscope.ProfileInuseSpace,
}

var profileTypes = map[string]pyroscope.ProfileType{
	string(pyroscope.ProfileCPU):           pyroscope.ProfileCPU,
	string(pyroscope.ProfileAllocObjects):  pyroscope.ProfileAllocObjects,
	string(pyroscope.ProfileAllocSpace):    pyroscope.ProfileAllocSpace,
	string(pyroscope.ProfileInuseObjects):  pyroscope.ProfileInuseObjects,
	string(pyroscope.ProfileInuseSpace):    pyroscope.ProfileInuseSpace,
	string(pyroscope.ProfileGoroutines):    pyroscope.ProfileGoroutines,
	string(pyroscope.ProfileMutexCount):    pyroscope.ProfileMutexCount,
	string(pyroscope.ProfileMutexDuration): pyroscope.ProfileMutexDuration,
	string(pyroscope.ProfileBlockCount):    pyroscope.ProfileBlockCount,
	string(pyroscope.ProfileBlockDuration): pyroscope.ProfileBlockDuration,
	string(pyroscope.ProfileGoroutineLeak): pyroscope.ProfileGoroutineLeak,
}

// Defaults defines service-owned profiling defaults.
type Defaults struct {
	ApplicationName string
	Tags            map[string]string
}

// StopFunc stops profiling and flushes remaining profile data.
type StopFunc func() error

// Start starts profiling when it is enabled in config.
func Start(conf config.Profiling, defaults Defaults) (StopFunc, error) {
	if !conf.Enabled {
		return func() error { return nil }, nil
	}

	applicationName := strings.TrimSpace(conf.ApplicationName)
	if applicationName == "" {
		applicationName = strings.TrimSpace(defaults.ApplicationName)
	}
	if applicationName == "" {
		return nil, fmt.Errorf("applicationName is empty")
	}

	serverAddress := strings.TrimSpace(conf.ServerAddress)
	if serverAddress == "" {
		return nil, fmt.Errorf("serverAddress is empty")
	}

	configuredProfileTypes, err := buildProfileTypes(conf.ProfileTypes)
	if err != nil {
		return nil, err
	}

	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName:   applicationName,
		ServerAddress:     serverAddress,
		BasicAuthUser:     conf.BasicAuthUser,
		BasicAuthPassword: conf.BasicAuthPassword,
		TenantID:          conf.TenantID,
		Tags:              buildTags(defaults.Tags, conf.Tags),
		ProfileTypes:      configuredProfileTypes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start pyroscope profiler: %w", err)
	}

	var once sync.Once
	var stopErr error
	return func() error {
		once.Do(func() {
			stopErr = profiler.Stop()
		})
		return stopErr
	}, nil
}

func buildProfileTypes(confTypes []string) ([]pyroscope.ProfileType, error) {
	if len(confTypes) == 0 {
		return append([]pyroscope.ProfileType(nil), defaultProfileTypes...), nil
	}

	builtTypes := make([]pyroscope.ProfileType, 0, len(confTypes))
	for _, confType := range confTypes {
		profileType, ok := profileTypes[strings.TrimSpace(confType)]
		if !ok {
			return nil, fmt.Errorf("unsupported profile type %q", confType)
		}
		builtTypes = append(builtTypes, profileType)
	}

	return builtTypes, nil
}

func buildTags(defaultTags map[string]string, confTags map[string]string) map[string]string {
	if len(defaultTags) == 0 && len(confTags) == 0 {
		return nil
	}

	tags := make(map[string]string, len(defaultTags)+len(confTags))
	maps.Copy(tags, defaultTags)
	maps.Copy(tags, confTags)

	return tags
}
