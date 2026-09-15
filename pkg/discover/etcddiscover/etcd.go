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
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultEtcdPrefix      = "/bknodemgr"
	defaultEtcdDialTimeout = 5 * time.Second
	defaultEtcdLeaseTTLSec = 10
	defaultListTickTime    = 10 * time.Second

	// defaultRetryInterval is the backoff before retrying a lost registration
	// or a broken watch.
	defaultRetryInterval = time.Second

	// defaultRevokeTimeout bounds revoking the lease of a stopped
	// registration. Revoking is an optimization over letting the orphaned
	// lease expire, so it must not block a shutdown on an unreachable etcd.
	defaultRevokeTimeout = 3 * time.Second

	etcdEntryBaseKeyValueCapacity = 8
)

var _ discover.IProvider = &ProviderEtcd{}

// ProviderEtcd implements discover.IProvider.
type ProviderEtcd struct {
	config *config.Etcd

	etcdClient     *clientv3.Client
	discoverPrefix string

	ctx    context.Context
	cancel context.CancelFunc

	serviceWatchList []discover.ServiceName

	// registrations of the instances registered by current runtime.
	registrationMutex sync.Mutex
	registrations     map[discover.ServiceName]map[string]*registration

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
		registrations:    make(map[discover.ServiceName]map[string]*registration),
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
	// cancel before stopping the registrations: cancelling aborts the etcd
	// requests a registration may have in flight, and that is what bounds the
	// wait for its lock below. Stopping first lets an unreachable etcd hold
	// the shutdown for as long as it stays unreachable.
	if provider.cancel != nil {
		provider.cancel()
	}

	// revoke the leases of the locally registered instances, dropping them
	// from discovery at once instead of after a lease TTL.
	for _, reg := range provider.takeAllRegistrations() {
		if err := reg.stop(); err != nil {
			logger.G.Sys().WithErr(err).
				With("service", reg.serviceName, "id", reg.instanceID).
				Warn("failed to deregister instance on shutdown")
		}
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

// Register registers service instances.
func (provider *ProviderEtcd) Register(serviceName discover.ServiceName, instances ...discover.Instance) error {
	// validate every instance before registering any of them, so that an
	// invalid input is reported without leaving a partial registration.
	for _, instance := range instances {
		if serviceName == "" {
			return fmt.Errorf(
				"failed to register instance, instance(%v): %w",
				instance,
				discover.ErrInvalidServiceName(),
			)
		}

		if err := instance.Validate(); err != nil {
			return fmt.Errorf("failed to register instance, instance(%v): %w", instance, err)
		}
	}

	if provider.etcdClient == nil {
		return discover.ErrDiscoverNotStarted()
	}

	for _, instance := range instances {
		if err := provider.register(serviceName, instance); err != nil {
			return fmt.Errorf("failed to register instance, instance(%v): %w", instance, err)
		}
	}

	return nil
}

// register registers a service instance and keeps it registered until it is
// deregistered or the provider is shut down.
func (provider *ProviderEtcd) register(serviceName discover.ServiceName, instance discover.Instance) error {
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

	key := provider.instanceKey(serviceName, instance.ID)

	// re-registering the same instance replaces the previous registration:
	// stop it before publishing the new lease, so that a write still pending
	// on it cannot rebind the key to the lease that is about to be revoked,
	// which would leave the key deleted with nothing to bring it back.
	if previous := provider.takeRegistration(serviceName, instance.ID); previous != nil {
		provider.stopReplacedRegistration(previous)
	}

	session, err := provider.newSession()
	if err != nil {
		return err
	}

	if err = provider.putInstance(key, instance, session.Lease()); err != nil {
		provider.closeSession(key, session)

		return err
	}

	reg := &registration{
		provider:    provider,
		serviceName: serviceName,
		instanceID:  instance.ID,
		key:         key,
		instance:    instance,
		session:     session,
	}

	// a concurrent Register of the same instance may have installed a
	// registration in the meantime: stop it so its keeper and lease are not
	// left behind.
	if previous := provider.putRegistration(reg); previous != nil {
		provider.stopReplacedRegistration(previous)
	}

	go reg.keep()

	logger.G.Sys().With("service", serviceName, "id", instance.ID, "data", instance).Info("registered service")

	return nil
}

// stopReplacedRegistration stops a registration that a newer one for the same
// instance has replaced.
func (provider *ProviderEtcd) stopReplacedRegistration(reg *registration) {
	if err := reg.stop(); err != nil {
		logger.G.Sys().WithErr(err).With("key", reg.key).Warn("failed to stop the replaced registration")
	}
}

// Update updates a service instance.
func (provider *ProviderEtcd) Update(serviceName discover.ServiceName, instance discover.Instance) error {
	if provider.etcdClient == nil {
		return discover.ErrDiscoverNotStarted()
	}

	reg := provider.getRegistration(serviceName, instance.ID)
	if reg == nil {
		return discover.ErrNotRegistered()
	}

	if err := reg.update(instance); err != nil {
		logger.G.Sys().WithErr(err).With("service", serviceName, "id", instance.ID).Error("failed to update instance")

		return err
	}

	logger.G.Sys().With("service", serviceName, "id", instance.ID, "data", instance).Info("successfully updated")

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

	reg := provider.takeRegistration(serviceName, instanceID)
	if reg == nil {
		return discover.ErrNotRegistered()
	}

	if err := reg.stop(); err != nil {
		return err
	}

	logger.G.Sys().With("service", serviceName, "id", instanceID).Info("successfully deregistered")

	return nil
}

// newSession creates an etcd session: a lease kept alive in the background,
// whose Done channel closes once the lease is no longer being refreshed.
func (provider *ProviderEtcd) newSession() (*concurrency.Session, error) {
	return concurrency.NewSession(
		provider.etcdClient,
		concurrency.WithContext(provider.ctx),
		concurrency.WithTTL(defaultEtcdLeaseTTLSec),
	)
}

// revokeSession ends a session and revokes its lease.
// The revoke runs on its own bounded context instead of provider.ctx: it must
// still reach etcd while shutting down, once that context is already
// cancelled, and it must never wait on an unreachable etcd. Orphaning already
// stops the keepalives, so a failed revoke only means the lease expires within
// a TTL rather than being dropped right now.
func (provider *ProviderEtcd) revokeSession(session *concurrency.Session) error {
	session.Orphan()

	ctx, cancel := context.WithTimeout(context.Background(), defaultRevokeTimeout)
	defer cancel()

	_, err := provider.etcdClient.Revoke(ctx, session.Lease())

	return err
}

// closeSession revokes the lease of a session that is no longer needed.
func (provider *ProviderEtcd) closeSession(key string, session *concurrency.Session) {
	if err := provider.revokeSession(session); err != nil {
		logger.G.Sys().WithErr(err).With("key", key).Warn("failed to close the unused session")
	}
}

// putInstance writes the instance to etcd, bound to the given lease.
func (provider *ProviderEtcd) putInstance(
	key string, instance discover.Instance, leaseID clientv3.LeaseID,
) error {

	content, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	_, err = provider.etcdClient.Put(provider.ctx, key, string(content), clientv3.WithLease(leaseID))

	return err
}

func (provider *ProviderEtcd) instanceKey(serviceName discover.ServiceName, instanceID string) string {
	return filepath.Join(provider.discoverPrefix, string(serviceName), instanceID)
}

// putRegistration stores the registration, returning the one it replaced.
func (provider *ProviderEtcd) putRegistration(reg *registration) *registration {
	provider.registrationMutex.Lock()
	defer provider.registrationMutex.Unlock()

	service, ok := provider.registrations[reg.serviceName]
	if !ok {
		service = make(map[string]*registration)
		provider.registrations[reg.serviceName] = service
	}

	previous := service[reg.instanceID]
	service[reg.instanceID] = reg

	return previous
}

// getRegistration returns the registration of an instance, nil when it is not
// registered by the current runtime.
func (provider *ProviderEtcd) getRegistration(
	serviceName discover.ServiceName, instanceID string,
) *registration {

	provider.registrationMutex.Lock()
	defer provider.registrationMutex.Unlock()

	return provider.registrations[serviceName][instanceID]
}

// takeRegistration removes and returns the registration of an instance, nil
// when it is not registered by the current runtime.
func (provider *ProviderEtcd) takeRegistration(
	serviceName discover.ServiceName, instanceID string,
) *registration {

	provider.registrationMutex.Lock()
	defer provider.registrationMutex.Unlock()

	reg, ok := provider.registrations[serviceName][instanceID]
	if !ok {
		return nil
	}

	delete(provider.registrations[serviceName], instanceID)

	return reg
}

// takeAllRegistrations removes and returns all the registrations.
func (provider *ProviderEtcd) takeAllRegistrations() []*registration {
	provider.registrationMutex.Lock()
	defer provider.registrationMutex.Unlock()

	regs := make([]*registration, 0)
	for serviceName, service := range provider.registrations {
		for _, reg := range service {
			regs = append(regs, reg)
		}

		delete(provider.registrations, serviceName)
	}

	return regs
}

// errRegistrationStopped is returned when the registration has been ended by
// Deregister or by a provider shutdown.
var errRegistrationStopped = errors.New("registration stopped")

// registration is a service instance registered by the current runtime.
// It owns the etcd session backing the instance key, and re-registers the
// instance whenever that session is lost.
type registration struct {
	provider    *ProviderEtcd
	serviceName discover.ServiceName
	instanceID  string
	key         string

	// mutex guards the registration state against concurrent Update,
	// Deregister and keeper goroutine access.
	mutex    sync.Mutex
	instance discover.Instance
	session  *concurrency.Session
	stopped  bool
}

// keep re-registers the instance whenever its etcd session is lost, until the
// instance is deregistered or the provider is shut down.
// Keeping a session alive does not write to etcd, so a healthy registration
// costs no etcd revision no matter how long it lives.
func (reg *registration) keep() {
	log := logger.G.Sys().With("service", reg.serviceName, "id", reg.instanceID)

	for {
		session := reg.currentSession()
		if session == nil {
			return
		}

		select {
		case <-reg.provider.ctx.Done():
			return

		case <-session.Done():
		}

		// the session is gone: its lease expired or the keepalive stream
		// died. Re-register to restore the instance key in etcd.
		err := reg.renew()
		if err == nil {
			log.Info("re-registered instance after session loss")

			continue
		}

		if errors.Is(err, errRegistrationStopped) {
			return
		}

		log.WithErr(err).Warn("failed to re-register instance, will retry")

		// the lost session is still the current one and its Done channel is
		// already closed, so back off before the next attempt.
		select {
		case <-reg.provider.ctx.Done():
			return

		case <-time.After(defaultRetryInterval):
		}
	}
}

// renew replaces the lost session with a fresh one and re-puts the instance.
func (reg *registration) renew() error {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	if reg.stopped {
		return errRegistrationStopped
	}

	session, err := reg.provider.newSession()
	if err != nil {
		return err
	}

	if err = reg.provider.putInstance(reg.key, reg.instance, session.Lease()); err != nil {
		reg.provider.closeSession(reg.key, session)

		return err
	}

	// release the lost session: its lease has either expired already or no
	// longer holds the instance key, which is now bound to the new lease.
	reg.session.Orphan()
	reg.session = session

	return nil
}

// update replaces the registered instance and pushes it to etcd.
func (reg *registration) update(instance discover.Instance) error {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	if reg.stopped {
		return discover.ErrNotRegistered()
	}

	// the stored instance is what the keeper re-registers with: store it
	// first, so that a put failing on a dying lease is still healed with the
	// new instance instead of the previous one.
	reg.instance = instance

	return reg.provider.putInstance(reg.key, instance, reg.session.Lease())
}

// stop ends the registration: it stops the keeper and revokes the session
// lease, which removes the instance key from etcd at once.
func (reg *registration) stop() error {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	if reg.stopped {
		return nil
	}
	reg.stopped = true

	return reg.provider.revokeSession(reg.session)
}

// currentSession returns the session currently backing the instance key, nil
// when the registration has been stopped.
func (reg *registration) currentSession() *concurrency.Session {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()

	if reg.stopped {
		return nil
	}

	return reg.session
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

// watch keeps a watch on a service running until the provider is shut down.
// A watch is cancelled by etcd on its own (a compaction outrunning the
// watcher, a stream error), so it is restarted instead of leaving the cache
// to the periodic list alone.
func (provider *ProviderEtcd) watch(serviceName discover.ServiceName) {
	logger.G.Sys().With("service", serviceName).Info("started watch for service")

	for {
		provider.watchOnce(serviceName)

		select {
		case <-provider.ctx.Done():
			logger.G.Sys().With("service", serviceName).Info("stopped watch for service")

			return

		case <-time.After(defaultRetryInterval):
		}
	}
}

// watchOnce consumes a single watch stream, returning when etcd closes it.
func (provider *ProviderEtcd) watchOnce(serviceName discover.ServiceName) {
	instanceHolder := provider.getCacheInstanceHolder(serviceName)

	rch := provider.etcdClient.Watch(
		provider.ctx,
		filepath.Join(provider.discoverPrefix, string(serviceName)),
		clientv3.WithPrefix())

	for wresp := range rch {
		if err := wresp.Err(); err != nil {
			logger.G.Sys().WithErr(err).With("service", serviceName).Warn("watch cancelled, will restart")

			return
		}

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
				instanceHolder.delete(id)

				logger.G.Sys().With("service", serviceName, "id", id).Info("observed instance delete")
			}
		}
	}
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

// instanceHolder is the read model of a service: the instances synced from
// etcd by the watch and list loops.
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
