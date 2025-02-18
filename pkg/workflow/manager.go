// Package workflow provides workflow manager.
package workflow

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/RichardKnop/machinery/v2"
	redisBackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisBroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	machineryConfig "github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/blog"
	"github.com/TencentBlueKing/bk-nodemgr/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Manager represents a workflow manager, provides workflow consumer and producer.
type Manager struct {
	// config
	config *ManagerConfig

	// state
	isRunning   bool
	isConsuming bool

	// context
	ctx    context.Context
	cancel context.CancelFunc

	server *machinery.Server
	worker *machinery.Worker

	storage Storage

	registeredActionDefs map[string]ActionDef

	launchWorkerErr chan error
}

type ManagerConfig struct {
	Redis   config.Redis
	MongoDB config.MongoDB

	WorkerNum int
}

// NewManager creates a new workflow manager.
func NewManager(config *ManagerConfig, storage Storage) *Manager {
	return &Manager{
		config:    config,
		isRunning: false,

		registeredActionDefs: make(map[string]ActionDef),

		storage: storage,

		launchWorkerErr: make(chan error, 1),
	}
}

// Start starts the workflow manager
func (m *Manager) Start(ctx context.Context) error {
	if m.isRunning {
		return errors.New("manager already started")
	}

	m.ctx, m.cancel = context.WithCancel(ctx)

	if err := m.RegisterAction(NewActionPeriodLauncher(m)); err != nil {
		return err
	}

	if err := m.initialize(ctx); err != nil {
		return err
	}

	if m.config.WorkerNum > 0 {
		if err := m.launchWorker(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) WaitWorkerShutdown() error {
	if !m.isRunning {
		return errors.New("manager is not running")
	}

	if !m.isConsuming {
		return errors.New("worker is not running")
	}

	return <-m.launchWorkerErr
}

// RegisterAction registers a workflow action.
func (m *Manager) RegisterAction(actionDef ActionDef) error {
	if m.isRunning {
		return errors.New("manager already started, can not register action")
	}

	if _, ok := m.registeredActionDefs[actionDef.Name()]; ok {
		return errors.New("action already registered")
	}

	m.registeredActionDefs[actionDef.Name()] = actionDef

	return nil
}

// RegisterActions registers multiple workflow actions.
func (m *Manager) RegisterActions(actionDefs ...ActionDef) error {
	for _, actionDef := range actionDefs {
		if err := m.RegisterAction(actionDef); err != nil {
			return err
		}
	}

	return nil
}

// GetRegisteredAction returns a registered workflow action.
func (m *Manager) GetRegisteredAction(name string) ActionDef {
	return m.registeredActionDefs[name]
}

func (m *Manager) DispatchTask(task *Task) error {
	if !m.isRunning {
		return errors.New("manager is not running")
	}

	if err := m.validateTask(task); err != nil {
		return err
	}

	if err := m.createTask(task); err != nil {
		return err
	}

	return m.dispatchTask(task)
}

func (m *Manager) StopTask(taskID string) error {
	return m.storage.MarkTaskStopping(taskID)
}

func (m *Manager) DispatchPeriodTask(task *Task) error {
	if !m.isRunning {
		return errors.New("manager is not running")
	}

	if err := m.validatePeriodTask(task); err != nil {
		return err
	}

	if err := m.createTask(task); err != nil {
		return err
	}

	return m.dispatchPeriodTask(task)
}

func (m *Manager) initialize(ctx context.Context) error {
	if m.config == nil {
		return errors.New("manager config is nil")
	}

	mongoClient, err := mongo.Connect(
		ctx,
		&mongoOptions.ClientOptions{
			Hosts: m.config.MongoDB.Hosts,
			Auth: &mongoOptions.Credential{
				Username:      m.config.MongoDB.Username,
				Password:      m.config.MongoDB.Password,
				AuthSource:    m.config.MongoDB.AuthSource,
				AuthMechanism: m.config.MongoDB.AuthMechanism,
			},
		})
	if err != nil {
		return err
	}
	mongoClient.Disconnect(ctx)

	mConfig := &machineryConfig.Config{
		DefaultQueue:    "bknodeman_tasks",
		ResultsExpireIn: 3600,
		Redis: &machineryConfig.RedisConfig{
			MaxIdle:                10,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}

	redisAddress := fmt.Sprintf("%s:%d", m.config.Redis.Host, m.config.Redis.Port)
	broker := redisBroker.New(mConfig, redisAddress, m.config.Redis.Password, "", 0)
	backend := redisBackend.New(mConfig, redisAddress, m.config.Redis.Password, "", 0)
	// lock := redisLock.New(mConfig, []string{fmt.Sprintf("%s@%s", m.config.Redis.Password, redisAddress)}, 0, 3)
	lock := eager.New()

	m.server = machinery.NewServer(mConfig, broker, backend, lock)

	m.isRunning = true

	return nil
}

func (m *Manager) launchWorker() error {
	if !m.isRunning {
		return errors.New("manager is not running")
	}

	for _, actionDef := range m.registeredActionDefs {
		if err := m.server.RegisterTask(actionDef.Name(), m.do); err != nil {
			return err
		}
	}

	m.worker = m.server.NewWorker("", m.config.WorkerNum)

	m.worker.SetErrorHandler(func(err error) {})
	m.worker.SetPreTaskHandler(func(signature *tasks.Signature) {})
	m.worker.SetPostTaskHandler(func(signature *tasks.Signature) {})

	m.isConsuming = true
	go func() {
		m.launchWorkerErr <- m.worker.Launch()
		m.isConsuming = false
	}()

	return nil
}

func (m *Manager) validateTask(task *Task) error {
	if task == nil {
		return errors.New("task is nil")
	}

	if task.pipeline.name == "" {
		return errors.New("pipeline name is empty")
	}

	return nil
}

func (m *Manager) validatePeriodTask(task *Task) error {
	if err := m.validateTask(task); err != nil {
		return err
	}

	if !task.data.IsPeriod {
		return fmt.Errorf("given task is not a period task")
	}

	return nil
}

func (m *Manager) dispatchTask(task *Task) error {
	var signatures []*tasks.Signature

	for _, actionDef := range task.pipeline.actionDefs {
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("%s-%s", task.data.TaskID, actionDef.Name()),
			Name: actionDef.Name(),
			Args: []tasks.Arg{
				{
					Name:  "action",
					Type:  "string",
					Value: actionDef.Name(),
				},
			},
		}

		signatures = append(signatures, signature)
	}

	if len(signatures) == 0 {
		return fmt.Errorf("no action to do")
	}

	// first action should be set task-id.
	signatures[0].Args = append(signatures[0].Args, tasks.Arg{
		Name:  "task_id",
		Type:  "string",
		Value: task.data.TaskID,
	})

	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		return err
	}

	_, err = m.server.SendChainWithContext(m.ctx, chain)
	if err != nil {
		return fmt.Errorf("send chain to machinery failed: %v", err)
	}

	return nil
}

func (m *Manager) dispatchPeriodTask(task *Task) error {
	var signatures []*tasks.Signature

	for _, actionDef := range task.pipeline.actionDefs {
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("%s-%s", task.data.TaskID, actionDef.Name()),
			Name: actionDef.Name(),
			Args: []tasks.Arg{
				{
					Name:  "action",
					Type:  "string",
					Value: actionDef.Name(),
				},
			},
		}

		signatures = append(signatures, signature)
	}

	if len(signatures) == 0 {
		return fmt.Errorf("no action to do")
	}

	// first action should be set task-id.
	signatures[0].Args = append(signatures[0].Args, tasks.Arg{
		Name:  "task_id",
		Type:  "string",
		Value: task.data.TaskID,
	})

	if err := m.server.RegisterPeriodicChain(task.data.PeriodSpec, "test", signatures...); err != nil {
		return fmt.Errorf("send chain to machinery failed: %v", err)
	}

	return nil
}

func (m *Manager) do(actionName string, taskID string) (string, error) {
	actionDef, ok := m.registeredActionDefs[actionName]
	if !ok {
		return taskID, fmt.Errorf("action %s not registered", actionName)
	}

	task, err := m.getTask(taskID)
	if err != nil {
		return taskID, err
	}

	action, err := task.Action(actionName)
	if err != nil {
		blog.Errorf("failed to get action data from task. action(%s), task-id(%s): %v", actionName, taskID, err)

		return taskID, err
	}

	if !task.data.IsPeriod {
		switch action.data.State {
		case ActionStateSuccess, ActionStateSkipped:
			blog.Infof("action is already done. action(%s), task-id(%s), state(%s)", actionName, taskID, action.data.State)

			return taskID, nil

		case ActionStateFailed, ActionStateTimeout, ActionStateStopped, ActionStateUnknown:
			blog.Errorf("failed to do action, action is not in pending state. action(%s), task-id(%s), state(%s)",
				actionName, taskID, action.data.State)

			return taskID, fmt.Errorf("task %s, action %s failed: %s", taskID, actionName, action.data.State)
		}
	}

	nowTime := time.Now()
	if action.data.Index == 0 {
		task.data.StartedAt = nowTime
	}
	action.data.StartedAt = nowTime
	action.data.State = ActionStateRunning

	if err = task.save(); err != nil {
		blog.Errorf("failed to save task data. action(%s), task-id(%s): %v", actionName, taskID, err)

		return taskID, err
	}

	actionCtx, actionCancel := context.WithTimeout(context.Background(), actionDef.Timeout())
	defer actionCancel()

	startedAt := task.data.StartedAt
	taskCtx, taskCancel := context.WithDeadline(context.Background(), startedAt.Add(task.data.Timeout))
	defer taskCancel()

	// watch storage for stopping event.
	stoppingC := m.storage.WatchTaskStopping(actionCtx, task.data.TaskID)

	doErr := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				doErr <- fmt.Errorf("action %s panic: %v, stack: %s", actionName, r, debug.Stack())
			}
		}()

		doErr <- actionDef.Do(&ActionContext{
			Ctx:    actionCtx,
			Action: action,
		})
	}()

	select {
	case err = <-doErr:
		blog.Infof("action done. action(%s), task-id(%s), err(%v)", actionName, taskID, err)

		if err != nil {
			action.data.State = ActionStateFailed
			action.data.EndedAt = time.Now()
			task.save()

			return taskID, err
		}

		action.data.State = ActionStateSuccess
		action.data.EndedAt = time.Now()
		task.save()

		return taskID, nil

	case <-actionCtx.Done():
		blog.Infof("action has timed out. action(%s), task-id(%s)", actionName, taskID)

		return taskID, fmt.Errorf("task %s, action %s timeout", taskID, actionName)

	case <-taskCtx.Done():
		blog.Infof("action has timed out, cause task has timed out. action(%s), task-id(%s)", actionName, taskID)

		return taskID, fmt.Errorf("task %s timeout", taskID)

	case <-stoppingC:
		blog.Infof("action has been stopped. action(%s), task-id(%s)", actionName, taskID)

		return taskID, fmt.Errorf("task %s has been stopped", taskID)
	}
	// TODO: graceful shutdown when manager context done.
}

func (m *Manager) createTask(task *Task) error {
	if err := m.storage.CreateTaskData(task.data); err != nil {
		blog.Errorf("failed to create task data, task-id(%s): %v", task.data.TaskID, err)

		return err
	}

	blog.Infof("successfully create task data, task-id(%s)", task.data.TaskID)
	return nil
}

func (m *Manager) getTask(taskID string) (*Task, error) {
	data, err := m.storage.GetTaskData(taskID)
	if err != nil {
		blog.Errorf("failed to get task data, task-id(%s): %v", taskID, err)

		return nil, err
	}

	pipeline := &Pipeline{name: data.Pipeline}
	for _, actionName := range data.Actions {
		actionDef, ok := m.registeredActionDefs[actionName]
		if !ok {
			blog.Errorf("failed to get task data, action not registered. task-id(%s), action(%s)", taskID, actionName)

			return nil, fmt.Errorf("action %s not registered", actionName)
		}

		pipeline = pipeline.Next(actionDef)
	}

	return &Task{
		data:       data,
		pipeline:   pipeline,
		saveMethod: m.storage.UpdateTaskData,
	}, nil
}
