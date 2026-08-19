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

// Package envx provides environment variables methods.
package envx

import (
	"fmt"
	"os"
	"strconv"
)

// LoadString get string value to store. Return true if successfully got from env.
func LoadString(key string, store *string) bool {
	if value, ok := os.LookupEnv(key); ok {
		*store = value

		return true
	}

	return false
}

// LoadInt get int value to store. Return true if successfully got from env.
func LoadInt(key string, store *int) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return false, nil
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return false, fmt.Errorf("failed to load int config from env, it is not a valid int. key(%s), value(%s)", key, value)
	}
	*store = result

	return true, nil
}

// LoadBool get bool value to store. Return true if successfully got from env.
func LoadBool(key string, store *bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok {

		return false, nil
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("failed to load bool config from env, it is not a valid int. key(%s), value(%s)", key, value)
	}
	*store = result

	return true, nil

}

// MustLoadString get string value to store.
func MustLoadString(key string, store *string) error {
	if !LoadString(key, store) {
		return fmt.Errorf("failed to load string config from env, key does not exist. key(%s)", key)
	}

	return nil
}

// MustLoadInt get int value to store.
func MustLoadInt(key string, store *int) error {
	ok, err := LoadInt(key, store)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("failed to load int config from env, key does not exist. key(%s)", key)
	}

	return nil
}

// MustLoadBool get bool value to store.
func MustLoadBool(key string, store *bool) error {
	ok, err := LoadBool(key, store)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("failed to load bool config from env, key does not exist. key(%s)", key)
	}

	return nil
}
