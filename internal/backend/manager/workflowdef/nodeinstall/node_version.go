/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package nodeinstall

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
)

// VersionParam defines the version param.
type VersionParam struct {
	daoRelease  release.IStorage
	ReleaseType types.ReleaseType
	Generation  types.Generation
	OSType      string
	CPUArch     string
	Version     string
}

func autoSelectVersion(ctx context.Context, versionParam VersionParam) (string, error) {
	platform, err := platform.Normalize(versionParam.OSType, versionParam.CPUArch)
	if err != nil {
		return "", fmt.Errorf("invalid platform, err: %w", err)
	}

	cond := buildReleaseCondition(versionParam, platform, "")

	releases, num, err := versionParam.daoRelease.ListRelease(ctx, types.UnlimitedPage(), cond)
	if err != nil {
		return "", fmt.Errorf("failed to list releases, err: %w", err)
	}
	if num == 0 {
		return "", fmt.Errorf("failed to list releases for platform. platform(%v)", platform)
	}

	defaultReleases := make([]string, 0)
	for _, release := range releases {
		if !release.AsDefault {
			continue
		}
		defaultReleases = append(defaultReleases, release.Version)
	}

	switch len(defaultReleases) {
	case 0:
		return "", fmt.Errorf("failed to get default release for platform. platform(%v)", platform)
	case 1:
		return defaultReleases[0], nil
	default:
		return "", fmt.Errorf(
			"failed to get default release for platform. platform(%v). multiple default releases found default-releases(%v)",
			platform, defaultReleases)
	}
}

func checkVersionAvailability(ctx context.Context, versionParam VersionParam) error {
	platform, err := platform.Normalize(string(versionParam.OSType), string(versionParam.CPUArch))
	if err != nil {
		return fmt.Errorf("invalid platform, err: %w", err)
	}

	cond := buildReleaseCondition(versionParam, platform, versionParam.Version)

	num, err := versionParam.daoRelease.CountRelease(ctx, cond)

	if err != nil {
		return fmt.Errorf("failed to check release version, err: %w", err)
	}

	if num == 0 {
		return fmt.Errorf(
			"failed to check release version. platform(%v). no release found", platform)
	}

	return nil
}

func buildReleaseCondition(versionParam VersionParam, plat platform.Platform, version string) *types.ReleaseCondition {
	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Type:       []types.ReleaseType{versionParam.ReleaseType},
			Platform:   []platform.Platform{plat},
			Generation: []types.Generation{versionParam.Generation},
		},
	}

	if version != "" {
		cond.ExactInclude.Version = []string{version}
	}

	return cond
}
