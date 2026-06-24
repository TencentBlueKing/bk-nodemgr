/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmdb

import (
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
)

const (
	// TopoNodeObjIDBiz topo node object id for biz.
	TopoNodeObjIDBiz = "biz"
	// TopoNodeObjIDHost topo node object id for host.
	TopoNodeObjIDHost = "host"
	// TopoNodeObjIDSet topo node object id for set.
	TopoNodeObjIDSet = "set"
	// TopoNodeObjIDModule topo node object id for module.
	TopoNodeObjIDModule = "module"
)

type iObjectKeeper interface {
	getName() string
	update(ctx contextx.IContext) error
}

func newSetObjectKeeper(cli *cli) *objectBasicKeeper {
	return newObjectBasicKeeper(cli, TopoNodeObjIDSet)
}

func newModuleObjectKeeper(cli *cli) *objectBasicKeeper {
	return newObjectBasicKeeper(cli, TopoNodeObjIDModule)
}

func newObjectBasicKeeper(cli *cli, objectID string) *objectBasicKeeper {
	return &objectBasicKeeper{
		cli:      cli,
		objectID: objectID,
	}
}

type objectBasicKeeper struct {
	cli      *cli
	objectID string

	mutex sync.RWMutex
	name  string
}

func (keeper *objectBasicKeeper) getName() string {
	keeper.mutex.RLock()
	defer keeper.mutex.RUnlock()

	return keeper.name
}

func (keeper *objectBasicKeeper) update(ctx contextx.IContext) error {
	resp, err := keeper.cli.searchObject(ctx, &SearchObjectReq{BKObjID: keeper.objectID})
	if err != nil {
		return err
	}

	for _, objectInfo := range *resp {
		if objectInfo == nil || objectInfo.BKObjectID != keeper.objectID {
			continue
		}

		keeper.mutex.Lock()
		keeper.name = objectInfo.BKObjectName
		keeper.mutex.Unlock()

		return nil
	}

	return fmt.Errorf("cmdb object not found: %s", keeper.objectID)
}
