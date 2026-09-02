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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	name    = "bk-nodemgr-deploy-policy-probe"
	version = "1.0.1"
)

func main() {
	configPath, debug, handled := parseArgs(os.Args[1:])
	if handled {
		return
	}
	if configPath == "" {
		usage()
		os.Exit(2)
	}

	if _, err := os.Stat(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stat config %s: %v\n", configPath, err)
		os.Exit(1)
	}

	run(configPath, debug)
}

func parseArgs(args []string) (string, bool, bool) {
	if len(args) == 0 {
		return "", false, false
	}

	var configPath string
	var debug bool
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--version", "version":
			fmt.Printf("%s %s\n", name, version)
			return "", false, true
		case "-c", "--config":
			if i+1 >= len(args) {
				usage()
				os.Exit(2)
			}
			i++
			configPath = args[i]
		case "--debug", "debug":
			debug = true
		case "status", "dump", "health":
			fmt.Printf("name=%s\n", name)
			fmt.Printf("version=%s\n", version)
			fmt.Println("pid_owner=gse")
			return "", false, true
		default:
			usage()
			os.Exit(2)
		}
	}

	return configPath, debug, false
}

func run(configPath string, debug bool) {
	if debug {
		fmt.Printf("%s debug started with config %s\n", name, configPath)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	for {
		sig := <-sigCh
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			if debug {
				fmt.Printf("%s debug stopped\n", name)
			}
			return
		case syscall.SIGUSR1:
			recordReload(configPath)
		}
	}
}

func recordReload(configPath string) {
	reloadPath := filepath.Join(filepath.Dir(configPath), name+".reload")
	content := []byte(fmt.Sprintf("%d\n", time.Now().Unix()))
	if err := os.WriteFile(reloadPath, content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to record reload: %v\n", err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s {-v|status|dump|health|-c CONFIG [--debug]}\n", name)
}
