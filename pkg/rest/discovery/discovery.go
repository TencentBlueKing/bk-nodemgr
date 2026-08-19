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

// Package discovery defines server discovery operations.
package discovery

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
)

// Interface discovery interface.
type Interface interface {
	// GetEndpoints get endpoints.
	GetEndpoints() ([]string, error)
}

// Discovery used to common discovery.
type Discovery struct {
	Name      string
	endpoints []string
	index     int
	sync.Mutex
}

// NewDiscovery create a discovery.
func NewDiscovery(name string, endpoints []string) Interface {
	return &Discovery{
		Name:      name,
		endpoints: endpoints,
		index:     0,
	}
}

// GetEndpoints get endpoints.
func (d *Discovery) GetEndpoints() ([]string, error) {
	d.Lock()
	defer d.Unlock()
	num := len(d.endpoints)
	if num == 0 {
		return []string{}, fmt.Errorf("there is no server can be used, name:(%s)", d.Name)
	}

	if d.index < num-1 {
		d.index++
		return append(d.endpoints[d.index-1:], d.endpoints[:d.index-1]...), nil
	}

	d.index = 0

	return append(d.endpoints[num-1:], d.endpoints[:num-1]...), nil
}

// ServiceDiscovery discovery with specific service in discover provider.
type ServiceDiscovery struct {
	discover     discover.Discover
	serviceName  discover.ServiceName
	endpointName discover.EndpointName
}

// NewServiceDiscovery create a service discovery.
func NewServiceDiscovery(
	discov discover.Discover, serviceName discover.ServiceName, endpointName discover.EndpointName) Interface {

	return &ServiceDiscovery{
		discover:     discov,
		serviceName:  serviceName,
		endpointName: endpointName,
	}
}

// GetEndpoints get endpoints.
func (sd *ServiceDiscovery) GetEndpoints() ([]string, error) {
	endpoints, err := sd.discover.GetAllEndpoint(sd.serviceName, sd.endpointName)
	if err != nil {
		return nil, err
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("there is no endpoint can be used. service(%s), endpoint(%s)",
			sd.serviceName, sd.endpointName)
	}

	data := make([]string, len(endpoints))
	for idx, ep := range endpoints {
		data[idx] = "http://" + net.JoinHostPort(ep.IPV4, strconv.Itoa(ep.Port))
	}

	return data, nil
}
