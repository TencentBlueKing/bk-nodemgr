/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package main ...
package main

import (
	"errors"
	"sync"
)

// nolint: gochecknoglobals
var downloadEndPoint = struct {
	sync.Once
	endpoint string
}{}

// SetDownloadEndPoint set download endpoint and callback endpoint.
func SetDownloadEndPoint(endpoint string) error {
	var err error
	downloadEndPoint.Do(func() {
		if endpoint == "" {
			err = errors.New("download endpoint is empty")
			return
		}

		downloadEndPoint.endpoint = endpoint
	})
	if err != nil {
		return err
	}

	return nil
}

// nolint: gochecknoglobals
var callbackEndPoint = struct {
	sync.Once
	endpoint string
}{}

// SetCallbackEndPoint set callback endpoint and callback endpoint.
func SetCallbackEndPoint(endpoint string) error {
	var err error
	callbackEndPoint.Do(func() {
		if endpoint == "" {
			err = errors.New("callback endpoint is empty")
			return
		}

		callbackEndPoint.endpoint = endpoint
	})
	if err != nil {
		return err
	}

	return nil
}

// GetDownloadEndPoint get download endpoint.
func GetDownloadEndPoint() string {
	return downloadEndPoint.endpoint
}

// GetCallBackEndpoint get callback endpoint.
func GetCallBackEndpoint() string {
	return callbackEndPoint.endpoint
}
