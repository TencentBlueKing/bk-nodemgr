/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package logger provides the global logger implements.
// nolint: gochecknoglobals,mnd
package logger

import (
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TencentBlueKing/bk-nodemgr/pkg/contextx"
	"go.opentelemetry.io/otel/trace"
)

// Level logger level.
// A message written to a high-severity log file is also written to each
// lower-severity log file.
type Level int32

const (
	// LevelDebug debug level.
	LevelDebug Level = iota

	// LevelInfo info level.
	LevelInfo

	// LevelWarn warn level.
	LevelWarn

	// LevelError error level.
	LevelError

	// LevelFatal fatal level.
	LevelFatal

	levelMax = 5
)

// get returns the value of the Level.
func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

// set sets the value of the Level.
func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

// Category logger category.
type Category int32

const (
	// CategorySystem system category.
	CategorySystem Category = iota

	// CategoryBusiness business category.
	CategoryBusiness

	categoryMax = 2
)

var (
	printer ILoggerPrinter = &logging

	onceInit sync.Once

	// G provides the global default logger.
	G ILogger = &Logger{Depth: 2}
)

// Config defines the logger config.
type Config struct {
	LogDir       string
	LogMaxSizeMB int
	LogMaxNum    int

	ToStdErr     bool
	AlsoToStdErr bool
	Level        Level
}

// Init initialize logger from config.
func Init(config Config) {
	onceInit.Do(func() {
		logDir = config.LogDir
		logMaxSize = uint64(config.LogMaxSizeMB) * 1024 * 1024
		logMaxNum = config.LogMaxNum

		logging.toStderr = config.ToStdErr
		logging.alsoToStderr = config.AlsoToStdErr
		logging.verbosity.set(config.Level)
	})
}

// Logger defines logger.
type Logger struct {
	// Depth defines the depth from caller.
	Depth int
}

// Sys creates sys logger.
func (l Logger) Sys() ILoggerOption {
	return &Option{
		category: CategorySystem,
		kvs:      make([]interface{}, 0),
		depth:    l.Depth,
	}
}

// Biz creates biz logger.
func (l Logger) Biz(ctx contextx.IContext) ILoggerOption {
	return &Option{
		ctx:      ctx,
		category: CategoryBusiness,
		kvs:      make([]interface{}, 0),
		depth:    l.Depth,
	}
}

// Flush flushes all logs into files.
func (l Logger) Flush() {
	logging.lockAndFlushAll()
}

// Option logger option.
type Option struct {
	category Category
	ctx      contextx.IContext
	kvs      []interface{}
	err      error
	duration *time.Duration
	level    Level
	depth    int
	assignFn func(format string, args ...interface{})
}

// Ctx sets context.
func (o *Option) Ctx(ctx contextx.IContext) ILoggerOption {
	o.ctx = ctx

	return o
}

// With adds some key-value pairs of context.
func (o *Option) With(kvs ...interface{}) ILoggerOption {
	o.kvs = append(o.kvs, kvs...)

	return o
}

// WithErr add error into logger.
func (o *Option) WithErr(err error) ILoggerOption {
	o.err = err

	return o
}

// WithDuration add duration into logger.
func (o *Option) WithDuration(duration time.Duration) ILoggerOption {
	o.duration = &duration

	return o
}

// AssignWhenLogging assigns the logger message to str when logging.
func (o *Option) AssignWhenLogging(str *string) ILoggerOption {
	if str != nil {
		o.assignFn = func(format string, args ...interface{}) {
			*str = fmt.Sprintf(format+o.additionMessage(), args...)
		}
	}

	return o
}

// Debug logs debug message.
func (o *Option) Debug(format string, args ...interface{}) {
	o.level = LevelDebug
	o.log(format, args...)
}

// Info logs info message.
func (o *Option) Info(format string, args ...interface{}) {
	o.level = LevelInfo
	o.log(format, args...)
}

// Warn logs warn message.
func (o *Option) Warn(format string, args ...interface{}) {
	o.level = LevelWarn
	o.log(format, args...)
}

// Error logs error message.
func (o *Option) Error(format string, args ...interface{}) {
	o.level = LevelError
	o.log(format, args...)
}

// DebugWriter returns the writer for debug message.
func (o *Option) DebugWriter() io.Writer {
	o.level = LevelDebug

	return Writer{o: o}
}

// InfoWriter returns the writer for info message.
func (o *Option) InfoWriter() io.Writer {
	o.level = LevelInfo

	return Writer{o: o}
}

// WarnWriter returns the writer for warn message.
func (o *Option) WarnWriter() io.Writer {
	o.level = LevelWarn

	return Writer{o: o}
}

// ErrorWriter returns the writer for error message.
func (o *Option) ErrorWriter() io.Writer {
	o.level = LevelError

	return Writer{o: o}
}

// Log messages.
func (o *Option) log(format string, args ...interface{}) {
	o.parseArgs()

	printer.Log(o.category, o.level, o.depth, format+o.additionMessage(), args...)

	if o.assignFn != nil {
		o.assignFn(format, args...)
	}
}

func (o *Option) additionMessage() string {
	var message string

	// there are key(value) pairs.
	for i := 0; i < (len(o.kvs)+1)/2; i++ {
		k := i * 2
		v := k + 1

		if i > 0 {
			message += ", "
		}

		if v >= len(o.kvs) {
			message += fmt.Sprintf("ignored key without value: %s", o.kvs[k])
			break
		}

		message += fmt.Sprintf("%s(%v)", o.kvs[k], o.kvs[v])
	}

	if message != "" {
		message = ". " + message
	}

	return message
}

func (o *Option) parseArgs() {
	if o.duration != nil {
		o.kvs = append(o.kvs, "cost", strconv.FormatInt(o.duration.Milliseconds(), 10)+"ms")
	}

	if o.ctx != nil {
		values := o.ctx.Values()
		data := make([]interface{}, 0, len(values)*2) // nolint: mnd
		for k, v := range values {
			data = append(data, k, v)
		}

		o.kvs = append(data, o.kvs...)
	}

	if o.err != nil {
		o.kvs = append(o.kvs, "err", o.err)
	}

	spanContext := trace.SpanContextFromContext(o.ctx)
	if spanContext.IsValid() {
		o.kvs = append(o.kvs, "trace-id", spanContext.TraceID().String())
		o.kvs = append(o.kvs, "span-id", spanContext.SpanID().String())
	}
}

// Writer provides a writer for logger.
type Writer struct {
	o *Option
}

// Write implements io.Writer.
func (w Writer) Write(p []byte) (n int, err error) {
	w.o.log(string(p))

	return len(p), nil
}
