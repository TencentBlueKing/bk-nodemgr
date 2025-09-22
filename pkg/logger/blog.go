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

// Package logger provides blog implements.
// nolint: varnamelen,gochecknoglobals,mnd,nestif,gochecknoinits
package logger

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	digits    = "0123456789"
	levelChar = "DIWEF"
)

var (
	levelName = []string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
	}

	categoryName = []string{
		CategorySystem:   "system",
		CategoryBusiness: "business",
	}
)

// buffer holds a byte Buffer for reuse. The zero value is ready for use.
type buffer struct {
	bytes.Buffer
	tmp  [64]byte // temporary byte array for creating headers.
	next *buffer
}

// twoDigits formats a zero-prefixed two-digit integer at buf.tmp[i].
func (buf *buffer) twoDigits(i, d int) {
	buf.tmp[i+1] = digits[d%10]
	d /= 10
	buf.tmp[i] = digits[d%10]
}

// nDigits formats an n-digit integer at buf.tmp[i],
// padding with pad on the left.
// It assumes d >= 0.
func (buf *buffer) nDigits(n, i, d int, pad byte) {
	j := n - 1
	for ; j >= 0 && d > 0; j-- {
		buf.tmp[i+j] = digits[d%10]
		d /= 10
	}
	for ; j >= 0; j-- {
		buf.tmp[i+j] = pad
	}
}

// someDigits formats a zero-prefixed variable-width integer at buf.tmp[i].
func (buf *buffer) someDigits(i, d int) int {
	// Print into the top, then copy down. We know there's space for at least
	// a 10-digit number.
	j := len(buf.tmp)
	for {
		j--
		buf.tmp[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}

	return copy(buf.tmp[i:], buf.tmp[j:])
}

var timeNow = time.Now // Stubbed out for testing.

// global logging.
var logging loggingT

// loggingT collects all the global state of the logging setup.
type loggingT struct {
	// print log to stderr.
	toStderr bool

	// print log to files and also to stderr.
	alsoToStderr bool

	// freeList is a list of byte buffers, maintained under freeListMu.
	freeList *buffer
	// freeListMu maintains the free list. It is separate from the main mutex
	// so buffers can be grabbed and printed to without holding the main lock,
	// for better parallelization.
	freeListMu sync.Mutex

	// mu protects the remaining elements of this structure and is
	// used to synchronize logging.
	mu sync.Mutex
	// file holds writer for each of the log types.
	file      [categoryMax][levelMax]flushSyncWriter
	verbosity Level
}

// Log print log as printer.
func (l *loggingT) Log(category Category, lv Level, depth int, format string, args ...interface{}) {
	if l.verbosity <= lv {
		l.printDepth(category, lv, depth, format, args...)
	}
}

func (l *loggingT) printDepth(category Category, lv Level, depth int, format string, args ...interface{}) {
	buf, _, _ := l.header(lv, depth)
	_, _ = fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(category, lv, buf, false)
}

// getBuffer returns a new, ready-to-use buffer.
func (l *loggingT) getBuffer() *buffer {
	l.freeListMu.Lock()
	b := l.freeList
	if b != nil {
		l.freeList = b.next
	}
	l.freeListMu.Unlock()
	if b == nil {
		b = new(buffer)
	} else {
		b.next = nil
		b.Reset()
	}

	return b
}

// putBuffer returns a buffer to the free list.
func (l *loggingT) putBuffer(b *buffer) {
	if b.Len() >= 256 {
		// Let big buffers die a natural death.
		return
	}
	l.freeListMu.Lock()
	b.next = l.freeList
	l.freeList = b
	l.freeListMu.Unlock()
}

/*
header formats a log header as defined by the C++ implementation.
It returns a buffer containing the formatted header and the user's file and line number.
The depth specifies how many stack frames above lives the source line to be identified in the log message.

Log lines have this form:

	Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...

where the fields are defined as follows:

	L                A single character, representing the log level (eg 'I' for INFO)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	file             The file name
	line             The line number
	msg              The user-supplied message
*/
func (l *loggingT) header(lv Level, depth int) (*buffer, string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}

	return l.formatHeader(lv, file, line), file, line
}

// formatHeader formats a log header using the provided file name and line number.
func (l *loggingT) formatHeader(lv Level, file string, line int) *buffer {
	now := timeNow()
	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}
	if lv > LevelFatal {
		lv = LevelInfo // for safety.
	}
	buf := l.getBuffer()

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	_, month, day := now.Date()
	hour, minute, second := now.Clock()
	// Lmmdd hh:mm:ss.uuuuuu threadid file:line]
	buf.tmp[0] = levelChar[lv]
	buf.twoDigits(1, int(month))
	buf.twoDigits(3, day)
	buf.tmp[5] = ' '
	buf.twoDigits(6, hour)
	buf.tmp[8] = ':'
	buf.twoDigits(9, minute)
	buf.tmp[11] = ':'
	buf.twoDigits(12, second)
	buf.tmp[14] = '.'
	buf.nDigits(6, 15, now.Nanosecond()/1000, '0')
	buf.tmp[21] = ' '
	buf.Write(buf.tmp[:22])

	buf.tmp[0] = ':'
	n := buf.someDigits(1, line)
	buf.tmp[n+1] = ']'
	buf.tmp[n+2] = ' '

	// add padding space
	logPointLength := len(file) + n
	for logPointLength < 20 {
		buf.WriteString(" ")
		logPointLength++
	}

	buf.WriteString(file)
	buf.Write(buf.tmp[:n+3])

	return buf
}

// output writes the data to the log files and releases the buffer.
func (l *loggingT) output(cg Category, lv Level, buf *buffer, alsoToStderr bool) {
	l.mu.Lock()
	data := buf.Bytes()
	leastLevel := l.verbosity.get()
	if l.toStderr {
		_, _ = os.Stderr.Write(data)
	} else {
		if alsoToStderr || l.alsoToStderr {
			_, _ = os.Stderr.Write(data)
		}
		if l.file[cg][lv] == nil {
			if err := l.createFiles(cg, lv); err != nil {
				_, _ = os.Stderr.Write(data) // Make sure the message appears somewhere.
				l.exit(err)
			}
		}

		for log := lv; log >= leastLevel; log-- {
			_, _ = l.file[cg][log].Write(data)
		}
	}
	if lv == LevelFatal {
		// Dump all goroutine stacks before exiting.
		// First, make sure we see the trace for the current goroutine on standard error.
		// If -logtostderr has been specified, the loop below will do that anyway
		// as the first stack in the full dump.
		if !l.toStderr {
			_, _ = os.Stderr.Write(stacks(false))
		}
		// Write the stack trace for all goroutines to the files.
		trace := stacks(true)
		logExitFunc = func(error) {} // If we get a write error, we'll still exit below.

		for log := LevelFatal; log >= leastLevel; log-- {
			if f := l.file[cg][log]; f != nil { // Can be nil if -logtostderr is set.
				_, _ = f.Write(trace)
			}
		}
		l.mu.Unlock()
		timeoutFlush(10 * time.Second)
		os.Exit(255) // C++ uses -1, which is silly because it's anded with 255 anyway.
	}
	l.putBuffer(buf)
	l.mu.Unlock()
}

// createFiles creates all the log files for level from sev down to debugLog.
// l.mu is held.
func (l *loggingT) createFiles(cg Category, level Level) error {
	now := time.Now()
	// AllFiles are created in decreasing severity order, so as soon as we find one
	// has already been created, we can stop.
	leastLevel := l.verbosity.get()
	for lv := level; lv >= leastLevel && l.file[cg][lv] == nil; lv-- {
		sb := &syncBuffer{
			logger:   l,
			category: cg,
			level:    lv,
		}
		if err := sb.rotateFile(now); err != nil {
			return err
		}
		l.file[cg][lv] = sb
	}

	return nil
}

const flushInterval = 30 * time.Second

// flushDaemon periodically flushes the log file buffers.
func (l *loggingT) flushDaemon() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	for range ticker.C {
		l.lockAndFlushAll()
	}
}

// lockAndFlushAll is like flushAll but locks l.mu first.
func (l *loggingT) lockAndFlushAll() {
	l.mu.Lock()
	l.flushAll()
	l.mu.Unlock()
}

// flushAll flushes all the logs and attempts to "sync" their data to disk.
// l.mu is held.
func (l *loggingT) flushAll() {
	// Flush from fatal down, in case there's trouble flushing.
	leastLevel := l.verbosity.get()
	for cg := range []Category{CategoryBusiness, CategorySystem} {
		for lv := LevelFatal; lv >= leastLevel; lv-- {
			file := l.file[cg][lv]
			if file != nil {
				_ = file.Flush() // ignore error
				_ = file.Sync()  // ignore error
			}
		}
	}
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}

	return trace
}

// timeoutFlush calls Flush and returns when it completes or after timeout
// elapses, whichever happens first.  This is needed because the hooks invoked
// by Flush may deadlock when glog.Fatal is called from a hook that holds
// a lock.
func timeoutFlush(timeout time.Duration) {
	done := make(chan bool, 1)
	go func() {
		logging.lockAndFlushAll()
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		_, _ = fmt.Fprintln(os.Stderr, "glog: Flush took longer than", timeout)
	}
}

// logExitFunc provides a simple mechanism to override the default behavior
// of exiting on error. Used in testing and to guarantee we reach a required exit
// for fatal logs. Instead, exit could be a function rather than a method but that
// would make its use clumsier.
var logExitFunc = func(error) {
	for cg := range []Category{CategoryBusiness, CategorySystem} {
		for _, f := range logging.file[cg] {
			// drop all data in buffer in case of buffer overflow.
			if f != nil {
				f.Reset()
			}
		}
	}
}

// exit is called if there is trouble creating or writing log files.
// It flushes the logs and exits the program; there's no point in hanging around.
// l.mu is held.
func (l *loggingT) exit(err error) {
	// If logExitFunc is set, we do that instead of exiting.
	if logExitFunc != nil {
		logExitFunc(err)
		return
	}

	_, _ = fmt.Fprintf(os.Stderr, "log: exiting because of error: %v\n", err)
	l.flushAll()
	os.Exit(2)
}

func init() {
	logging.toStderr = false
	logging.alsoToStderr = false
	logging.verbosity.set(LevelFatal)
	go logging.flushDaemon()
}
