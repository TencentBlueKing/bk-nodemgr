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

// File I/O for logs.

// Package logger provides file operations.
// nolint: varnamelen,gochecknoglobals,mnd,nestif,gochecknoinits,nonamedreturns,gocognit
package logger

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// flushSyncWriter is the interface satisfied by logging destinations.
type flushSyncWriter interface {
	Flush() error
	Sync() error
	Reset()
	io.Writer
}

// bufferSize sizes the buffer associated with each log file. It's large
// so that log records can accumulate without the logging thread blocking
// on disk I/O. The flushDaemon will block instead.
const bufferSize = 256 * 1024

// syncBuffer joins a bufio.Writer to its underlying file, providing access to the
// file's Sync method and providing a wrapper for the Write method that provides log
// file rotation. There are conflicting methods, so the file cannot be embedded.
// l.mu is held for all its methods.
type syncBuffer struct {
	logger *loggingT
	*bufio.Writer
	file     *os.File
	category Category
	level    Level
	nbytes   uint64 // The number of bytes written to this file
}

// Reset buffer.
func (sb *syncBuffer) Reset() {
	sb.Writer.Reset(sb.file)
}

// Sync buffer.
func (sb *syncBuffer) Sync() error {
	return sb.file.Sync()
}

// Write buffer.
func (sb *syncBuffer) Write(p []byte) (n int, err error) {
	if sb.nbytes+uint64(len(p)) >= MaxSize() {
		if err = sb.rotateFile(time.Now()); err != nil {
			sb.logger.exit(err)
		}
	}
	n, err = sb.Writer.Write(p)
	sb.nbytes += uint64(n)
	if err != nil {
		sb.logger.exit(err)
	}

	return n, err
}

// rotateFile closes the syncBuffer's file and starts a new one.
func (sb *syncBuffer) rotateFile(now time.Time) error {
	if sb.file != nil {
		_ = sb.Flush()
		_ = sb.file.Close()
	}
	var err error
	sb.file, _, err = create(categoryName[sb.category], levelName[sb.level], now)
	sb.nbytes = 0
	if err != nil {
		return err
	}

	sb.Writer = bufio.NewWriterSize(sb.file, bufferSize)

	// Write header.
	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "Log file created at: %s\n", now.Format("2006/01/02 15:04:05"))
	_, _ = fmt.Fprintf(&buf, "Running on machine: %s\n", host)
	_, _ = fmt.Fprintf(&buf, "Binary: Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS,
		runtime.GOARCH)
	_, _ = fmt.Fprintf(&buf, "Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid file:line] msg\n")
	n, err := sb.file.Write(buf.Bytes())
	sb.nbytes += uint64(n)

	return err
}

// logMaxSize is the maximum size of a log file in bytes.
var logMaxSize uint64 = 500 * 1024 * 1024

// MaxSize get log max size.
func MaxSize() uint64 {
	return logMaxSize
}

// logMaxNum is the maximum of log files for one thread.
var logMaxNum = 10

// MaxNum get log max num.
func MaxNum() int {
	return logMaxNum
}

// fileInfo contains log filename and its timestamp.
type fileInfo struct {
	name      string
	timestamp string
}

// fileInfoList implements Interface interface in sort. For
// sorting a list of fileInfo.
type fileInfoList []fileInfo

// Len for sorting.
func (b fileInfoList) Len() int { return len(b) }

// Swap for sorting.
func (b fileInfoList) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// Less for sorting.
func (b fileInfoList) Less(i, j int) bool { return b[i].timestamp < b[j].timestamp }

// fileBlock is a block of chain in logKeeper.
type fileBlock struct {
	fileInfo
	next *fileBlock
}

// logKeeper maintains a chain of each level log file. Its head
// is the earliest file while its tail is the oldest. It remains
// up to MaxNum() files, and the extra added will lead to delete
// the oldest. At first it load from logDir and take existing files
// into the chain. And remove the part over MaxNum().
type logKeeper struct {
	dir      string
	onceLoad sync.Once
	head     map[string]*fileBlock
	tail     map[string]*fileBlock
	total    map[string]int
}

func (lk *logKeeper) add(category, tag string, newBlock *fileBlock) (ok bool) {
	key := lk.key(category, tag)

	block, ok := lk.tail[key]
	if !ok {
		return
	}
	if block == nil {
		lk.head[key] = newBlock
	} else {
		if block.name == newBlock.name {
			return false
		}
		block.next = newBlock
	}
	lk.tail[key] = newBlock
	lk.total[key]++
	for lk.total[key] > MaxNum() {
		lk.remove(category, tag)
	}

	return ok
}

func (lk *logKeeper) remove(category, tag string) (ok bool) {
	key := lk.key(category, tag)

	block, ok := lk.head[key]
	if !ok || lk.total[key] == 0 {
		return
	}

	if block == nil {
		return
	}

	if err := lk.removeFile(block.name); err != nil {
		// 不能使用log输出，否则会死锁问题
		// log.Printf("remove file '%s' failed: %s", block.name, err.Error())
		fmt.Printf("remove file failed, block-name(%s): %v\n", block.name, err)
	}
	lk.head[key] = block.next
	block = nil // nolint: wastedassign
	lk.total[key]--

	return ok
}

func (lk *logKeeper) key(category, tag string) string {
	return category + ":" + tag
}

func (lk *logKeeper) removeFile(name string) error {
	return os.Remove(filepath.Join(lk.dir, name))
}

// dirMode means this dir can be read and write.
const dirMode = 0750

// load this func will load all files in logDir and add to logKeeper.
func (lk *logKeeper) load() {
	_, err := os.Stat(lk.dir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Printf("read dir failed: %v", err)
		return
	}

	if err = os.MkdirAll(lk.dir, dirMode); err != nil {
		fmt.Printf("mkdir failed: %v", err)
		return
	}

	tmpDir, err := os.ReadDir(lk.dir)
	if err != nil {
		fmt.Printf("read dir failed: %v", err)
		return
	}

	reg := logNameReg()
	tmp := make(map[string]fileInfoList)
	for _, fi := range tmpDir {
		if !fi.Type().IsRegular() {
			continue
		}

		result := reg.FindStringSubmatch(fi.Name())
		if result == nil {
			continue
		}

		name, category, timestamp, tag := result[0], result[1], result[2], result[3]
		key := lk.key(category, tag)

		if tmp[key] == nil {
			tmp[key] = make([]fileInfo, 0, len(tmpDir))
		}
		tmp[key] = append(tmp[key], fileInfo{name: name, timestamp: timestamp})
	}

	for key, blockList := range tmp {
		sort.Sort(blockList)
		for i, block := range blockList {
			if i <= MaxNum() {
				fb := &fileBlock{
					fileInfo: fileInfo{name: block.name, timestamp: block.timestamp},
					next:     nil,
				}
				if i == 0 {
					lk.head[key] = fb
				} else {
					lk.tail[key].next = fb
				}
				lk.tail[key] = fb
				lk.total[key]++
			} else {
				if err = lk.removeFile(block.name); err != nil {
					fmt.Printf("remove file failed, filename(%s): %v", block.name, err)
				}
			}
		}
	}
}

// logDirs lists the candidate directories for new log files.
var logDirs []*logKeeper

// If non-empty, overrides the choice of directory in which to write logs.
// See createLogDirs for the full list of possible destinations.
var logDir = "./logs"

func createLogDirs() {
	var dirs []string
	if logDir != "" {
		dirs = append(dirs, logDir)
	}
	dirs = append(dirs, os.TempDir())

	for _, dir := range dirs {
		head := make(map[string]*fileBlock)
		tail := make(map[string]*fileBlock)
		total := make(map[string]int)
		for _, name := range levelName {
			head[name] = nil
			tail[name] = nil
			total[name] = 0
		}

		logDirs = append(logDirs, &logKeeper{dir: dir, head: head, tail: tail, total: total})
	}
}

var (
	program  = filepath.Base(os.Args[0])
	host     = "unknownhost"
	userName = "unknownuser"
)

func init() {
	h, err := os.Hostname()
	if err == nil {
		host = shortHostname(h)
	}

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}

	// Sanitize userName since it may contain filepath separators on Windows.
	userName = strings.ReplaceAll(userName, `\`, "_")
}

// shortHostname returns its argument, truncating at the first period.
// For instance, given "www.google.com" it returns "www".
func shortHostname(hostname string) string {
	if i := strings.Index(hostname, "."); i >= 0 {
		return hostname[:i]
	}

	return hostname
}

// logName returns a new log file name containing tag, with start time t, and
// the name for the symlink for tag.
func logName(category, tag string, t time.Time) (string, string) {
	name := fmt.Sprintf("%s-%s.%04d%02d%02d%02d%02d%02d.%s",
		program,
		category,
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		tag)

	return name, program + "-" + category + "." + tag
}

// logNameReg returns a regexp object for match log file name.
func logNameReg() *regexp.Regexp {
	reg, _ := regexp.Compile(fmt.Sprintf(`^%s-(%s)\.(\d{8}\d{6})\.(%s)$`,
		program,
		strings.Join(categoryName, "|"),
		strings.Join(levelName, "|")))

	return reg
}

var onceLogDirs sync.Once

// create creates a new log file and returns the file and its filename, which
// contains tag ("INFO", "FATAL", etc.) and t.  If the file is created
// successfully, create also attempts to update the symlink for that tag, ignoring
// errors.
func create(category, tag string, t time.Time) (f *os.File, filename string, err error) {
	onceLogDirs.Do(createLogDirs)
	if len(logDirs) == 0 {
		return nil, "", errors.New("no log dirs")
	}
	name, link := logName(category, tag, t)
	var lastErr error
	for _, lk := range logDirs {
		lk.onceLoad.Do(lk.load)
		fname := filepath.Join(lk.dir, name)
		f, err := os.Create(fname) // nolint: gosec
		if err == nil {
			symlink := filepath.Join(lk.dir, link)
			_ = os.Remove(symlink)        // ignore err
			_ = os.Symlink(name, symlink) // ignore err
			lk.add(category, tag, &fileBlock{fileInfo: fileInfo{name: name, timestamp: ""}, next: nil})

			return f, fname, nil
		}
		lastErr = err
	}

	return nil, "", fmt.Errorf("cannot create log: %v", lastErr)
}
