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

// Jwt token generator.
//
// Build:
//
//	go build -o jwt-generator .
//
// Usage:
//
//	./jwt-generator -k <key> -e <expire-time>
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	restheader "github.com/TencentBlueKing/bk-nodemgr/pkg/rest/header"
)

func main() {
	var (
		key        string
		expireTime int64
		bkUserName string
		loginName  string
	)

	flag.StringVar(&key, "k", "", "Key")
	flag.Int64Var(&expireTime, "e", 0, "Expire Time")
	flag.StringVar(&bkUserName, "u", "admin", "Bk Username")
	flag.StringVar(&loginName, "l", "admin", "Login Name")
	flag.Parse()

	if key == "" {
		fmt.Fprintln(os.Stderr, "Error: key is required")
		os.Exit(1)
	}

	nodemgrAuthManager := restheader.NewNodeMgrAuthorizationManager(key, restheader.WithJwtTokenExpiration(time.Duration(expireTime)*time.Hour))
	token, err := nodemgrAuthManager.Generate(loginName, bkUserName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to generate token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(token)
}
