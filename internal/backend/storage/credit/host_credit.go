/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package credit

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/internal/backend/storage"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"github.com/google/uuid"
)

func generateHostCreditID() string {
	return fmt.Sprintf("host_%d_%s", time.Now().Unix(), uuid.New().String())
}

// CreateHostCredit create host credit.
func (s *Storage) CreateHostCredit(nCtx contextx.IContext, creditData []byte, expiredAt time.Time) (string, error) {
	var encryptedCreditData []byte
	var err error

	// record metric.
	metric := s.metric().Start("host_create")
	defer metric.End(err)

	if encryptedCreditData, err = s.crypter.Encrypt(creditData); err != nil {
		return "", err
	}

	creditID := generateHostCreditID()
	if err = s.daoCredit.Upsert(nCtx, creditID, encryptedCreditData, expiredAt); err != nil {
		return "", err
	}

	return creditID, nil
}

// LoadHostCredit load host credit.
func (s *Storage) LoadHostCredit(nCtx contextx.IContext, creditID string) ([]byte, error) {
	var encryptedCreditData, creditData []byte
	var err error

	// record metric.
	metric := s.metric().Start("host_load")
	defer metric.End(err)

	if encryptedCreditData, err = s.daoCredit.Get(nCtx, creditID); err != nil {
		return nil, err
	}

	if creditData, err = s.crypter.Decrypt(encryptedCreditData); err != nil {
		return nil, err
	}

	return creditData, nil
}

// CheckHostCreditValid check host credit valid.
func (s *Storage) CheckHostCreditValid(nCtx contextx.IContext, creditIDList ...string) (map[string]bool, error) {
	var result map[string]bool
	var err error

	// record metric.
	metric := s.metric().Start("host_check_valid")
	defer metric.End(err)

	if result, err = s.daoCredit.CheckValid(nCtx, creditIDList...); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Storage) metric() *storage.MetricData {
	return storage.Metric(StorageName)
}
