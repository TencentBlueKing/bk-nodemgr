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

package credit

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/google/uuid"
)

const (
	metricOperationCreateHostCredit     = "host_create"
	metricOperationLoadHostCredit       = "host_load"
	metricOperationCheckHostCreditValid = "host_check_valid"
)

func generateHostCreditID() string {
	return fmt.Sprintf("host_%d_%s", time.Now().Unix(), uuid.New().String())
}

// CreateHostCredit create host credit.
func (s *Storage) CreateHostCredit(nCtx contextx.IContext, creditData []byte, expiredAt time.Time) (string, error) {
	var (
		creditID            string
		encryptedCreditData []byte
		err                 error
	)

	err = s.WrapFn(nCtx, metricOperationCreateHostCredit, func(nCtx contextx.IContext) error {
		var err error
		if encryptedCreditData, err = s.crypter.Encrypt(creditData); err != nil {
			return err
		}

		creditID = generateHostCreditID()
		if err = s.daoCredit.Upsert(nCtx, creditID, encryptedCreditData, expiredAt); err != nil {
			return err
		}

		return nil
	})

	return creditID, err
}

// LoadHostCredit load host credit.
func (s *Storage) LoadHostCredit(nCtx contextx.IContext, creditID string) ([]byte, error) {
	var (
		encryptedCreditData []byte
		creditData          []byte
		err                 error
	)

	err = s.WrapFn(nCtx, metricOperationLoadHostCredit, func(nCtx contextx.IContext) error {
		var err error
		if encryptedCreditData, err = s.daoCredit.Get(nCtx, creditID); err != nil {
			return err
		}

		if creditData, err = s.crypter.Decrypt(encryptedCreditData); err != nil {
			return err
		}

		return nil
	})

	return creditData, err
}

// CheckHostCreditValid check host credit valid.
func (s *Storage) CheckHostCreditValid(nCtx contextx.IContext, creditIDList ...string) (map[string]bool, error) {
	var (
		result map[string]bool
		err    error
	)

	err = s.WrapFn(nCtx, metricOperationCheckHostCreditValid, func(nCtx contextx.IContext) error {
		var err error
		if result, err = s.daoCredit.CheckValid(nCtx, creditIDList...); err != nil {
			return err
		}

		return nil
	})

	return result, err
}
