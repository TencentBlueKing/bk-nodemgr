/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package filedownloader this package provide file download methods.
package filedownloader

import (
	"context"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/constant"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/logger"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/retrier"
	"github.com/TencentBlueKing/bk-nodemgr/tools/pkg/gopool"
)

// Step download files step.
type Step struct {
	downloadPoint        string
	callbackEndpoint     string
	pkgGeneration        int
	pkgPath              string
	pkgVersion           string
	nodeRole             constant.NodeRole
	token                string
	tmpAgentConfPath     string
	tmpFileProxyConfPath string
	tmpDataProxyConfPath string
	checkListPath        string
}

// StepArgs define args for step.
type StepArgs struct {
	DownloadPoint        string
	CallbackEndpoint     string
	PkgGeneration        int
	PkgPath              string
	PkgVersion           string
	NodeRole             constant.NodeRole
	Token                string
	TmpAgentConfPath     string
	TmpFileProxyConfPath string
	TmpDataProxyConfPath string
	CheckListPath        string
}

// NewStep ...
func NewStep(args StepArgs) *Step {
	step := &Step{
		downloadPoint:        args.DownloadPoint,
		callbackEndpoint:     args.CallbackEndpoint,
		pkgGeneration:        args.PkgGeneration,
		pkgPath:              args.PkgPath,
		pkgVersion:           args.PkgVersion,
		nodeRole:             args.NodeRole,
		token:                args.Token,
		tmpAgentConfPath:     args.TmpAgentConfPath,
		tmpFileProxyConfPath: args.TmpFileProxyConfPath,
		tmpDataProxyConfPath: args.TmpDataProxyConfPath,
		checkListPath:        args.CheckListPath,
	}

	return step
}

// Run run the step to download files.
// nolint: funlen,gocognit
func (step *Step) Run(ctx context.Context) error {
	logger.Infof(constant.StepDownloadFiles, "start to download files.")

	gp := gopool.NewPool()
	backoff := retrier.NewExpoBackoff(retrier.ExpoBackoffOptsDefault())

	// download gse config files.
	gp.Go(func() error {
		err := backoff.Do(ctx, func(attempt int) error {
			err := GetAgentConfig(ctx,
				step.callbackEndpoint,
				step.tmpAgentConfPath,
				string(step.nodeRole),
				step.token)
			if err != nil {
				logger.Errorf(constant.StepDownloadFiles,
					"get agent config failed, attempt: %d, err: %v", attempt, err)

				return fmt.Errorf("get agent config failed: %v", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		logger.Infof(constant.StepDownloadFiles, "successfully get agent config file.")

		return nil
	})

	if step.nodeRole == constant.NodeRoleProxy {
		gp.Go(func() error {
			err := backoff.Do(ctx, func(attempt int) error {
				if err := GetFileProxyConf(ctx, step.tmpFileProxyConfPath,
					string(step.nodeRole),
					step.token,
					step.callbackEndpoint); err != nil {
					logger.Errorf(constant.StepDownloadFiles,
						"get gse file proxy config failed, attempt: %d, err: %v", attempt, err)

					return fmt.Errorf("get gse file proxy config failed: %v", err)
				}

				return nil
			})
			if err != nil {
				return err
			}

			logger.Infof(constant.StepDownloadFiles, "successfully get gse file proxy config.")

			return nil
		})

		gp.Go(func() error {
			err := backoff.Do(ctx, func(attempt int) error {
				if err := GetDataProxyConf(ctx,
					step.tmpDataProxyConfPath,
					string(step.nodeRole),
					step.token,
					step.callbackEndpoint); err != nil {
					logger.Errorf(constant.StepDownloadFiles,
						"get gse data proxy config failed, attempt: %d, err: %v", attempt, err)

					return fmt.Errorf("get gse data proxy config failed: %v", err)
				}

				return nil
			})
			if err != nil {
				return err
			}

			logger.Infof(constant.StepDownloadFiles, "successfully get gse data proxy config.")

			return nil
		})
	}

	// download gse pkg.
	gp.Go(func() error {
		err := backoff.Do(ctx, func(attempt int) error {
			err := DownloadPkg(ctx,
				step.pkgGeneration,
				step.pkgPath,
				step.pkgVersion,
				step.downloadPoint,
				string(step.nodeRole))
			if err != nil {
				logger.Errorf(constant.StepDownloadFiles,
					"download agent pkg failed, attempt: %d, err: %v", attempt, err)

				return fmt.Errorf("download files failed: %v", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		logger.Infof(constant.StepDownloadFiles, "successfully download agent pkg.")

		return nil
	})

	// download check list.
	gp.Go(func() error {
		err := backoff.Do(ctx, func(attempt int) error {
			if err := GetCheckList(ctx,
				string(step.nodeRole),
				step.checkListPath,
				step.token,
				step.callbackEndpoint); err != nil {
				logger.Errorf(constant.StepDownloadFiles,
					"download check list failed, attempt: %d, err: %v", attempt, err)

				return fmt.Errorf("download check list failed: %v", err)
			}

			return nil
		})
		if err != nil {
			return err
		}

		logger.Infof(constant.StepDownloadFiles, "successfully download check list.")

		return nil
	})

	if err := gp.Wait(); err != nil {
		logger.Infof(constant.StepDownloadFiles, "failed to download files, err: %v", err)
		return errors.New("failed to download files")
	}

	logger.Infof(constant.StepDownloadFiles, "download files done.")

	return nil
}
