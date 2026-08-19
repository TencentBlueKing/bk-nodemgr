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

package step

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/flag"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/handler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/cmd/installer/node/persistent"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/agenthandler"
	"github.com/TencentBlueKing/bk-nodemgr/tools/internal/installer/node/precheck"
	"github.com/spf13/cobra"
)

// NewPreCheck creates a new precheck step command.
// nolint: lll
func NewPreCheck() *cobra.Command {
	var (
		// optional flags.
		preCheckListConf string

		// pre-run.
		agentHandler   agenthandler.IAgentHandler
		checkListParam precheck.CheckList
	)

	stepCmd := &cobra.Command{
		Use:   "precheck",
		Short: "Precheck",
		Long:  "Precheck",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			vars, err := persistent.GetVariables(cmd)
			if err != nil {
				return err
			}

			if preCheckListConf == "" {
				preCheckListConf = filepath.Join(vars.DataDir, "precheck.json")
			}

			agentHandler = handler.NewAgentHandler(vars.NodeRole, vars.DeployDir, vars.DeployEnv)

			// nolint: gosec
			content, err := os.ReadFile(preCheckListConf)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(content, &checkListParam); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			step := precheck.NewStep(precheck.StepArgs{
				AgentHandler: agentHandler,
				CheckList:    checkListParam,
			})

			if err := step.Run(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("successfully done precheck")

			return nil
		},
	}

	/*
	 * optional flags.
	 */
	stepCmd.Flags().StringVar(&preCheckListConf, flag.PreCheckListConf, "", "precheck list config file. if not set, will use the default data path")

	return stepCmd
}
