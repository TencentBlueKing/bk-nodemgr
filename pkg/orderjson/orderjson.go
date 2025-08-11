/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package orderjson provides the json ordered marshal/unmarshal.
package orderjson

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OrderedData is a json data can be marshalled and unmarshalled in ordered.
type OrderedData struct {
	orderedKeys []string
	dataMap     map[string]any

	dataSlice []any

	data any
}

// Set sets the value.
func (o *OrderedData) Set(key string, value any) error {
	if o.dataMap == nil {
		o.dataMap = make(map[string]any)
	}

	if o.orderedKeys == nil {
		o.orderedKeys = make([]string, 0)
	}

	if _, ok := o.dataMap[key]; !ok {
		o.orderedKeys = append(o.orderedKeys, key)
	}

	o.dataMap[key] = value

	return nil
}

// Get gets the value.
func (o *OrderedData) Get(key string) (any, error) {
	if o.dataMap == nil {
		return nil, fmt.Errorf("not found")
	}

	value, ok := o.dataMap[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return value, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *OrderedData) UnmarshalJSON(raw []byte) error {
	// check json is valid.
	var tmpValue any
	err := json.Unmarshal(raw, &tmpValue)
	if err != nil {
		return err
	}

	rawStr := strings.Trim(string(raw), " \t\n\r")
	o.orderedKeys = make([]string, 0)

	switch rawStr[0] {
	case '{':
		o.dataMap = make(map[string]any)

		itemStr := strings.Trim(rawStr[1:len(rawStr)-1], " \t\n\r")
		for len(itemStr) > 0 {
			key, next := cutdownFirstKey(itemStr)
			if next < 0 {
				return fmt.Errorf("not a valid ordered json data")
			}

			// record the ordered keys.
			o.orderedKeys = append(o.orderedKeys, key)

			var subNext int
			left := strings.Trim(itemStr[next:], " :\t\n\r")
			switch left[0] {
			case '{', '[':
				subNext = cutdownFirstElement(left)
			default:
				subNext = cutdownBeforeFirstComma(left)
				if subNext < 0 {
					subNext = len(left)
				}
			}

			if subNext < 0 {
				return fmt.Errorf("not a valid ordered json data")
			}

			subData := new(OrderedData)
			err := json.Unmarshal([]byte(left[:subNext]), subData)
			if err != nil {
				return err
			}

			o.dataMap[key] = subData
			itemStr = strings.Trim(left[subNext:], " ,\t\n\r")

			continue
		}

		return nil

	case '[':
		o.dataSlice = make([]any, 0)

		itemStr := strings.Trim(rawStr[1:len(rawStr)-1], " \t\n\r")
		for len(itemStr) > 0 {
			var next int
			switch itemStr[0] {
			case '{', '[':
				next = cutdownFirstElement(itemStr)
				if next < 0 {
					return fmt.Errorf("not a valid ordered json data")
				}
			default:
				next = cutdownBeforeFirstComma(itemStr)
				if next < 0 {
					next = len(itemStr)
				}
			}

			subData := new(OrderedData)
			err := json.Unmarshal([]byte(itemStr[:next]), subData)
			if err != nil {
				return err
			}

			o.dataSlice = append(o.dataSlice, subData)

			itemStr = strings.Trim(itemStr[next:], " ,\t\n\r")
		}

		return nil

	default:
		o.data = tmpValue

		return nil
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (o OrderedData) MarshalJSON() ([]byte, error) {
	if o.data != nil {
		return json.Marshal(o.data)
	}

	if o.dataSlice != nil {
		return json.Marshal(o.dataSlice)
	}

	if o.dataMap == nil {
		return nil, fmt.Errorf("not a valid ordered json data")
	}

	// build json.
	var b strings.Builder
	b.WriteString("{")

	for i, key := range o.orderedKeys {
		if i > 0 {
			b.WriteString(",")
		}

		// marshal key.
		keyBytes, _ := json.Marshal(key)
		b.Write(keyBytes)
		b.WriteString(":")

		// marshal value.
		valBytes, err := json.Marshal(o.dataMap[key])
		if err != nil {
			return nil, err
		}
		b.Write(valBytes)
	}

	b.WriteString("}")
	return []byte(b.String()), nil
}

func cutdownFirstKey(raw string) (string, int) {
	key := make([]rune, 0)
	skip := false
	inQuote := false
	for i, c := range raw {
		if skip {
			skip = false
			key = append(key, c)
			continue
		}

		if c == ' ' {
			continue
		}

		if c == '\\' {
			skip = true
			continue
		}

		if c == '"' {
			if inQuote {
				return string(key), i + 1
			}

			inQuote = true
			continue
		}

		if inQuote {
			key = append(key, c)
			continue
		}
	}

	return "", -1
}

func cutdownBeforeFirstComma(raw string) int {
	skip := false
	inQuote := false
	for i, c := range raw {
		if skip {
			skip = false
			continue
		}

		if c == ' ' {
			continue
		}

		if c == '\\' {
			skip = true
			continue
		}

		if c == '"' {
			inQuote = !inQuote
			continue
		}

		if !inQuote && c == ',' {
			return i
		}
	}

	return -1
}

func cutdownFirstElement(raw string) int {
	stack := make([]rune, 0)
	skip := false
	inQuote := false
	for i, c := range raw {
		if skip {
			skip = false
			continue
		}

		if c == ' ' {
			continue
		}

		if c == '\\' {
			skip = true
			continue
		}

		if c == '"' {
			inQuote = !inQuote
			continue
		}

		if inQuote {
			continue
		}

		switch c {
		case '{', '}', '[', ']':
			stack = append(stack, c)
		}

		for len(stack) >= 2 {
			switch string(stack[len(stack)-2:]) {
			case "{}", "[]":
				stack = stack[:len(stack)-2]
				continue
			}

			break
		}

		if len(stack) == 0 {
			return i + 1
		}
	}

	return -1
}
