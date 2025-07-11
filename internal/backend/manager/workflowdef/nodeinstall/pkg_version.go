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
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage/release"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/format/platform"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/criteria"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/types"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/workflow/action"
)

// VersionParam defines the version param.
type VersionParam struct {
	daoRelease  release.IStorage
	ReleaseType types.ReleaseType
	Generation  types.Generation
	OSType      criteria.OSType
	CPUArch     criteria.CPUArch
	Version     string
}

func autoSelectVersion(ctx *action.InstanceContext, versionParam VersionParam) (string, error) {
	p, err := platform.Normalize(string(versionParam.OSType), string(versionParam.CPUArch))
	if err != nil {
		return "", fmt.Errorf("invalid platform, err: %w", err)
	}

	cond := buildReleaseCondition(versionParam, p, "")

	releases, num, err := versionParam.daoRelease.ListRelease(ctx.Ctx, types.UnlimitedPage(), cond)
	if err != nil {
		return "", fmt.Errorf("failed to list releases, err: %w", err)
	}
	if num == 0 {
		return "", fmt.Errorf("no release found for platform(%s)", p)
	}

	var defaultReleases []string
	for _, release := range releases {
		if release.AsDefault {
			defaultReleases = append(defaultReleases, release.Version)
		}
	}

	switch len(defaultReleases) {
	case 0:
		return "", fmt.Errorf("no default release found for platform(%s)", p)
	case 1:
		return defaultReleases[0], nil
	default:
		return "", fmt.Errorf(
			"multiple default releases found for platform(%s),versions(%v)",
			p, defaultReleases)
	}
}

func checkVersionAvailability(ctx *action.InstanceContext, versionParam VersionParam) error {
	p, err := platform.Normalize(string(versionParam.OSType), string(versionParam.CPUArch))
	if err != nil {
		return fmt.Errorf("invalid platform, err: %w", err)
	}

	cond := buildReleaseCondition(versionParam, p, versionParam.Version)

	num, err := versionParam.daoRelease.CountRelease(ctx.Ctx, cond)

	if err != nil {
		return fmt.Errorf("failed to get release, err: %w", err)
	}

	if num == 0 {
		return fmt.Errorf("version(%s) not found", versionParam.Version)
	}

	return nil
}

func buildReleaseCondition(versionParam VersionParam, p platform.Platform, version string) *types.ReleaseCondition {
	cond := &types.ReleaseCondition{
		ExactInclude: &types.ReleaseExactFields{
			Type:       []types.ReleaseType{versionParam.ReleaseType},
			Platform:   []platform.Platform{p},
			Generation: []types.Generation{versionParam.Generation},
		},
	}

	if version != "" {
		cond.ExactInclude.Version = []string{version}
	}

	return cond
}
