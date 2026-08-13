/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-节点管理(BlueKing-BK-NODEMAN) available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Package etcddiscover implements a discover provider for etcd.
package etcddiscover

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/discover"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/logger"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/runtime/ssl"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultEtcdPrefix      = "/bknodemgr"
	defaultEtcdDialTimeout = 5 * time.Second
	defaultEtcdLeaseTTLSec = 10
	defaultListTickTime    = 10 * time.Second

	metaKeyLeaseID = "etcd-lease-id"

	etcdEntryBaseKeyValueCapacity = 8
)

// ProviderEtcd implements discover.IProvider.
type ProviderEtcd struct {
	config *config.Etcd

	etcdClient     *clientv3.Client
	discoverPrefix string

	ctx    context.Context
	cancel context.CancelFunc

	serviceWatchList []discover.ServiceName

	// local instances registered by current runtime.
	localInstancesMutex sync.RWMutex
	localInstances      map[discover.ServiceName]*instanceHolder

	// cache instances synced from etcd.
	cacheInstanceMutex sync.RWMutex
	cacheInstances     map[discover.ServiceName]*instanceHolder
}

// NewProviderEtcd creates a new ProviderEtcd.
func NewProviderEtcd(config *config.Etcd, opts ...OptionFn) *ProviderEtcd {
	provider := &ProviderEtcd{
		config:           config,
		discoverPrefix:   defaultEtcdPrefix,
		serviceWatchList: make([]discover.ServiceName, 0),
		localInstances:   make(map[discover.ServiceName]*instanceHolder),
		cacheInstances:   make(map[discover.ServiceName]*instanceHolder),
	}

	for _, opt := range opts {
		opt(provider)
	}

	return provider
}

// OptionFn is an option function for ProviderEtcd.
type OptionFn func(provider *ProviderEtcd)

// WithWatch adds services to watch.
func WithWatch(services ...discover.ServiceName) OptionFn {
	return func(provider *ProviderEtcd) {
		provider.serviceWatchList = append(provider.serviceWatchList, services...)
	}
}

// WithDiscoverPathPrefix sets the discover prefix.
func WithDiscoverPathPrefix(prefix string) OptionFn {
	return func(provider *ProviderEtcd) {
		provider.discoverPrefix = prefix
	}
}

type etcdLogFunc func(level logger.Level, msg string, kvs ...interface{})

type etcdLoggerCore struct {
	log    etcdLogFunc
	fields []zapcore.Field
}

func (core *etcdLoggerCore) Enabled(zapcore.Level) bool {
	return true
}

func (core *etcdLoggerCore) With(fields []zapcore.Field) zapcore.Core {
	clonedFields := append(append([]zapcore.Field{}, core.fields...), fields...)

	return &etcdLoggerCore{
		log:    core.log,
		fields: clonedFields,
	}
}

func (core *etcdLoggerCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !core.Enabled(entry.Level) {
		return checked
	}

	return checked.AddCore(entry, core)
}

func (core *etcdLoggerCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if core.log == nil {
		return nil
	}

	allFields := append(append([]zapcore.Field{}, core.fields...), fields...)
	core.log(mapEtcdLogLevel(entry.Level), entry.Message, etcdEntryKeyValues(entry, allFields)...)

	return nil
}

func (core *etcdLoggerCore) Sync() error {
	return nil
}

func newEtcdClientConfig(conf *config.Etcd, tlsConf *tls.Config) clientv3.Config {
	return clientv3.Config{
		Endpoints:   conf.Endpoints,
		Username:    conf.Username,
		Password:    conf.Password,
		DialTimeout: defaultEtcdDialTimeout,
		TLS:         tlsConf,
		Logger:      newEtcdLogger(),
	}
}

func newEtcdLogger() *zap.Logger {
	return zap.New(&etcdLoggerCore{log: writeEtcdLog}, zap.AddCaller())
}

func writeEtcdLog(level logger.Level, msg string, kvs ...interface{}) {
	logOption := logger.G.Sys().With(kvs...)

	switch level {
	case logger.LevelDebug:
		logOption.Debug(msg)
	case logger.LevelInfo:
		logOption.Info(msg)
	case logger.LevelWarn:
		logOption.Warn(msg)
	default:
		logOption.Error(msg)
	}
}

func mapEtcdLogLevel(level zapcore.Level) logger.Level {
	switch level {
	case zapcore.DebugLevel:
		return logger.LevelDebug
	case zapcore.InfoLevel:
		return logger.LevelInfo
	case zapcore.WarnLevel:
		return logger.LevelWarn
	default:
		return logger.LevelError
	}
}

func etcdEntryKeyValues(entry zapcore.Entry, fields []zapcore.Field) []interface{} {
	kvs := make([]interface{}, 0, len(fields)*2+etcdEntryBaseKeyValueCapacity)
	kvs = append(kvs, "level", entry.Level.String())

	if entry.Caller.Defined {
		kvs = append(kvs, "caller", entry.Caller.TrimmedPath())
	}
	if entry.LoggerName != "" {
		kvs = append(kvs, "logger", entry.LoggerName)
	}
	if entry.Stack != "" {
		kvs = append(kvs, "stack", entry.Stack)
	}

	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(encoder)
	}

	keys := make([]string, 0, len(encoder.Fields))
	for key := range encoder.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		kvs = append(kvs, key, encoder.Fields[key])
	}

	return kvs
}

// Start starts the provider.
func (provider *ProviderEtcd) Start(ctx context.Context) error {
	tlsConf, err := provider.initTLS()
	if err != errNotTLS && err != nil {
		return err
	}

	provider.etcdClient, err = clientv3.New(newEtcdClientConfig(provider.config, tlsConf))
	if err != nil {
		return err
	}

	provider.ctx, provider.cancel = context.WithCancel(ctx)

	// start service watching.
	provider.startWatching()

	// keep listing.
	go provider.keepListing()

	logger.G.Sys().Info("started etcd discover provider")

	return nil
}

// GracefulShutdown gracefully shuts down the provider, stop all activities.
func (provider *ProviderEtcd) GracefulShutdown() error {
	if provider.cancel != nil {
		provider.cancel()
	}

	logger.G.Sys().Info("stopped etcd discover provider")

	return nil
}

// GetAllService get all specific service instances.
func (provider *ProviderEtcd) GetAllService(serviceName discover.ServiceName) ([]discover.Instance, error) {
	return provider.getCacheInstanceHolder(serviceName).all(), nil
}

// GetAllEndpoint get all specific service endpoints.
func (provider *ProviderEtcd) GetAllEndpoint(
	serviceName discover.ServiceName, endpointName discover.EndpointName) ([]discover.Endpoint, error) {

	return provider.getCacheInstanceHolder(serviceName).getEndpoints(endpointName), nil
}

// GetEndpoint get specific service endpoint.
func (provider *ProviderEtcd) GetEndpoint(
	serviceName discover.ServiceName, endpointName discover.EndpointName, selector discover.Selector) (
	discover.Endpoint, error) {

	if selector == nil {
		return discover.Endpoint{}, discover.ErrInvalidSelector()
	}

	return provider.getCacheInstanceHolder(serviceName).getEndpoint(endpointName, selector)
}

// SelectEndpoints select a specified number of endpoints.
func (provider *ProviderEtcd) SelectEndpoints(
	serviceName discover.ServiceName, endpointName discover.EndpointName, count int, selector discover.Selector) ([]discover.Endpoint, error) {

	endpoints, err := provider.GetAllEndpoint(serviceName, endpointName)
	if err != nil {
		return nil, err
	}

	return discover.SelectEndpoints(endpoints, count, selector)
}

// Register registers a service instance.
// nolint: gocognit
func (provider *ProviderEtcd) Register(serviceName discover.ServiceName, instance discover.Instance) error {
	logger.G.Sys().With("service", serviceName, "id", instance.ID).Info("registering service")

	if serviceName == "" {
		return discover.ErrInvalidServiceName()
	}

	if instance.ID == "" {
		return discover.ErrInvalidInstanceID()
	}

	if instance.Name == "" {
		return discover.ErrInvalidInstanceName()
	}

	resp, err := provider.etcdClient.Grant(provider.ctx, defaultEtcdLeaseTTLSec)
	if err != nil {
		return err
	}

	leaseID := int64(resp.ID)
	instance.SetMeta(metaKeyLeaseID, leaseID)
	content, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	_, err = provider.etcdClient.Put(
		provider.ctx,
		filepath.Join(provider.discoverPrefix, string(serviceName), instance.ID),
		string(content),
		clientv3.WithLease(clientv3.LeaseID(leaseID)),
	)
	if err != nil {
		return err
	}

	ch, err := provider.etcdClient.KeepAlive(provider.ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		return err
	}

	go func() {
		for resp := range ch {
			logger.G.Sys().With("lease-id", resp.ID).Debug("recved grant keepalive response")
		}
		logger.G.Sys().With("lease-id", leaseID).Info("grant keepalive channel closed, goroutine exit")
	}()

	provider.getLocalInstanceHolder(serviceName).upsert(instance)
	logger.G.Sys().With("service", serviceName, "id", instance.ID, "data", string(content)).Info("registered service")

	// register keeper.
	go func() {
		for {
			select {
			case <-provider.ctx.Done():
				return
			default:
				cachedInstance, err := provider.getLocalInstanceHolder(serviceName).get(instance.ID)
				if err != nil {
					logger.G.Sys().WithErr(err).With("service", serviceName, "id", instance.ID).Warn("local instance not found, quit the register keeper")

					return
				}

				if err = provider.putService(serviceName, cachedInstance); err != nil {
					logger.G.Sys().WithErr(err).Warn("failed to put service in register keeper")
				}
			}

			time.Sleep(time.Second)
		}
	}()

	return nil
}

// Update updates a service instance.
func (provider *ProviderEtcd) Update(serviceName discover.ServiceName, instance discover.Instance) error {
	if err := provider.putService(serviceName, instance); err != nil {
		logger.G.Sys().WithErr(err).With("service", serviceName, "id", instance.ID).Error("failed to update instance")

		return err
	}

	logger.G.Sys().With("service", serviceName, "id", instance.ID, "data", instance).Info("successfully updated")

	return nil
}

func (provider *ProviderEtcd) putService(serviceName discover.ServiceName, instance discover.Instance) error {
	if provider.etcdClient == nil {
		return discover.ErrDiscoverNotStarted()
	}

	if serviceName == "" {
		return discover.ErrInvalidServiceName()
	}

	if instance.ID == "" {
		return discover.ErrInvalidInstanceID()
	}

	if instance.Name == "" {
		return discover.ErrInvalidInstanceName()
	}

	holder := provider.getLocalInstanceHolder(serviceName)
	instanceOld, err := holder.get(instance.ID)
	if err != nil {
		return err
	}

	leaseID, err := instanceOld.GetMetaInt64(metaKeyLeaseID)
	if err != nil {
		return errors.Join(discover.ErrDiscoverInternalError(), fmt.Errorf("lease id not found: %w", err))
	}

	instance.SetMeta(metaKeyLeaseID, leaseID)
	content, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	key := filepath.Join(provider.discoverPrefix, string(serviceName), instance.ID)
	if _, err = provider.etcdClient.Put(provider.ctx, key, string(content), clientv3.WithLease(clientv3.LeaseID(leaseID))); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Warn("failed to put service to etcd, need to grant new lease")

		resp, err := provider.etcdClient.Grant(provider.ctx, defaultEtcdLeaseTTLSec)
		if err != nil {
			return err
		}

		leaseID := int64(resp.ID)
		instance.SetMeta(metaKeyLeaseID, leaseID)

		if _, err = provider.etcdClient.Put(provider.ctx, key, string(content), clientv3.WithLease(clientv3.LeaseID(leaseID))); err != nil {
			return err
		}

		ch, err := provider.etcdClient.KeepAlive(provider.ctx, clientv3.LeaseID(leaseID))
		if err != nil {
			return err
		}

		go func() {
			for resp := range ch {
				logger.G.Sys().With("lease-id", resp.ID).Debug("recved grant keepalive response")
			}
			logger.G.Sys().With("lease-id", leaseID).Info("grant keepalive channel closed, goroutine exit")
		}()

		logger.G.Sys().With("key", key).Info("successfully grant new lease and update resource")
	}

	holder.upsert(instance)

	return nil
}

// Deregister deregisters a service instance.
func (provider *ProviderEtcd) Deregister(serviceName discover.ServiceName, instanceID string) error {
	if provider.etcdClient == nil {
		return discover.ErrDiscoverNotStarted()
	}

	if serviceName == "" {
		return discover.ErrInvalidServiceName()
	}

	if instanceID == "" {
		return discover.ErrInvalidInstanceID()
	}

	holder := provider.getLocalInstanceHolder(serviceName)
	instance, err := holder.get(instanceID)
	if err != nil {
		return err
	}

	leaseID, err := instance.GetMetaInt64(metaKeyLeaseID)
	if err != nil {
		return err
	}

	_, err = provider.etcdClient.Revoke(provider.ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		return err
	}

	holder.delete(instanceID)
	logger.G.Sys().With("service", serviceName, "id", instance.ID).Info("successfully deregistered")

	return nil
}

var errNotTLS = errors.New("not tls")

func (provider *ProviderEtcd) initTLS() (*tls.Config, error) {
	if provider.config.TLS.CAFile == "" || provider.config.TLS.CertFile == "" || provider.config.TLS.KeyFile == "" {
		return nil, errNotTLS
	}

	tlsConf := &ssl.TLSConfig{
		CAFile:   provider.config.TLS.CAFile,
		CertFile: provider.config.TLS.CertFile,
		KeyFile:  provider.config.TLS.KeyFile,
		Password: provider.config.TLS.Password,
	}

	if err := tlsConf.Validate(); err != nil {
		return nil, err
	}

	return tlsConf.NewClientTLSConf()
}

func (provider *ProviderEtcd) startWatching() {
	for _, serviceName := range provider.serviceWatchList {
		go provider.watch(serviceName)
	}
}

func (provider *ProviderEtcd) watch(serviceName discover.ServiceName) {
	logger.G.Sys().With("service", serviceName).Info("started watch for service")

	instanceHolder := provider.getCacheInstanceHolder(serviceName)

	rch := provider.etcdClient.Watch(
		provider.ctx,
		filepath.Join(provider.discoverPrefix, string(serviceName)),
		clientv3.WithPrefix())

	for wresp := range rch {
		for _, ev := range wresp.Events {
			id := filepath.Base(string(ev.Kv.Key))

			switch ev.Type {
			case clientv3.EventTypePut:
				var instance discover.Instance
				if err := json.Unmarshal(ev.Kv.Value, &instance); err != nil {
					logger.G.Sys().
						WithErr(err).
						With("service", serviceName, "id", instance.ID, "data", string(ev.Kv.Value)).
						Error("observed instance put, failed to unmarshal")

					continue
				}

				if id != instance.ID {
					logger.G.Sys().
						With("service", serviceName, "id", instance.ID, "data", string(ev.Kv.Value)).
						Error("observed instance put, id mismatch")

					continue
				}

				instanceHolder.upsert(instance)

				logger.G.Sys().
					With("service", serviceName, "id", instance.ID, "data", string(ev.Kv.Value)).
					Debug("observed instance put")

			case clientv3.EventTypeDelete:
				instanceHolder := provider.getCacheInstanceHolder(serviceName)
				instanceHolder.delete(id)

				logger.G.Sys().With("service", serviceName, "id", id).Info("observed instance delete")
			}
		}
	}

	logger.G.Sys().With("service", serviceName).Info("stopped watch for service")
}

func (provider *ProviderEtcd) keepListing() {
	logger.G.Sys().Info("started keep listing")

	// list first time.
	for _, serviceName := range provider.serviceWatchList {
		provider.list(serviceName)
	}

	ticker := time.NewTicker(defaultListTickTime)
	for {
		select {
		case <-provider.ctx.Done():
			logger.G.Sys().Info("stopped keep listing")
			return

		case <-ticker.C:
			for _, serviceName := range provider.serviceWatchList {
				provider.list(serviceName)
			}
		}
	}
}

func (provider *ProviderEtcd) list(serviceName discover.ServiceName) {
	resp, err := provider.etcdClient.Get(
		provider.ctx,
		filepath.Join(provider.discoverPrefix, string(serviceName)),
		clientv3.WithPrefix(),
	)
	if err != nil {
		logger.G.Sys().WithErr(err).With("service", serviceName).Error("failed to list instances")

		return
	}

	instanceHolder := provider.getCacheInstanceHolder(serviceName)

	validIDs := make([]string, 0)
	for _, kv := range resp.Kvs {
		id := filepath.Base(string(kv.Key))

		var instance discover.Instance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			logger.G.Sys().
				WithErr(err).
				With("service", serviceName, "id", instance.ID, "data", string(kv.Value)).
				Error("failed to unmarshal instance")

			continue
		}

		if id != instance.ID {
			logger.G.Sys().WithErr(err).Error("list instance but id mismatch. service(%s), id(%s), data(%s)",
				serviceName, id, string(kv.Value))

			continue
		}

		instanceHolder.upsert(instance)

		validIDs = append(validIDs, id)
	}

	instanceHolder.deleteNotIn(validIDs)
}

func (provider *ProviderEtcd) getCacheInstanceHolder(serviceName discover.ServiceName) *instanceHolder {
	provider.cacheInstanceMutex.Lock()
	defer provider.cacheInstanceMutex.Unlock()

	if holder, ok := provider.cacheInstances[serviceName]; ok {
		return holder
	}

	holder := &instanceHolder{
		instances: make(map[string]discover.Instance),
	}
	provider.cacheInstances[serviceName] = holder

	return holder
}

func (provider *ProviderEtcd) getLocalInstanceHolder(serviceName discover.ServiceName) *instanceHolder {
	provider.localInstancesMutex.Lock()
	defer provider.localInstancesMutex.Unlock()

	if holder, ok := provider.localInstances[serviceName]; ok {
		return holder
	}

	holder := &instanceHolder{
		instances: make(map[string]discover.Instance),
	}
	provider.localInstances[serviceName] = holder

	return holder
}

type instanceHolder struct {
	mutex     sync.RWMutex
	instances map[string]discover.Instance
}

func (holder *instanceHolder) upsert(instance discover.Instance) {
	holder.mutex.Lock()
	defer holder.mutex.Unlock()

	holder.instances[instance.ID] = instance
}

func (holder *instanceHolder) delete(ids ...string) {
	holder.mutex.Lock()
	defer holder.mutex.Unlock()

	for _, id := range ids {
		delete(holder.instances, id)
	}
}

func (holder *instanceHolder) deleteNotIn(ids []string) {
	needDeleteIDs := make([]string, 0)

	holder.mutex.Lock()
	for idInCache := range holder.instances {
		found := false
		for _, id := range ids {
			if idInCache == id {
				found = true

				break
			}
		}

		if !found {
			needDeleteIDs = append(needDeleteIDs, idInCache)
		}
	}
	holder.mutex.Unlock()

	holder.delete(needDeleteIDs...)
}

func (holder *instanceHolder) all() []discover.Instance {
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()

	instances := make([]discover.Instance, len(holder.instances))
	idx := 0
	for _, instance := range holder.instances {
		instances[idx] = instance
		idx++
	}

	return instances
}

func (holder *instanceHolder) get(id string) (discover.Instance, error) {
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()

	instance, ok := holder.instances[id]
	if !ok {
		return discover.Instance{}, discover.ErrNotRegistered()
	}

	return instance, nil
}

func (holder *instanceHolder) getEndpoints(endpointName discover.EndpointName) []discover.Endpoint {
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()

	endpoints := make([]discover.Endpoint, 0)
	for _, instance := range holder.instances {
		if endpoint, ok := instance.Endpoints[endpointName]; ok {
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

func (holder *instanceHolder) getEndpoint(endpointName discover.EndpointName, selector discover.Selector) (
	discover.Endpoint, error) {

	return selector.Select(holder.getEndpoints(endpointName))
}
