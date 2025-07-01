// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package winpath provides utilities for manipulating Windows-style file paths.
package winpath

import (
	"errors"
	"strings"
)

const (
	// DirSeparator is the path separator for Windows.
	DirSeparator = '\\'

	// ListSeparator is the path list separator for Windows.
	ListSeparator = ';'

	// DriverSeparator is the path separator for Windows drive letters.
	DriverSeparator = ':'
)

// Join joins any number of path elements into a single path,
// separating them with an OS specific DirSeparator. Empty elements
// are ignored. The result is Cleaned. However, if the argument
// list is empty or all its elements are empty, Join returns
// an empty string.
// On Windows, the result will only be a UNC path if the first
// non-empty element is a UNC path.
// nolint: varnamelen
func Join(elem ...string) string {
	var b strings.Builder
	var lastChar byte
	for _, e := range elem {
		switch {
		case b.Len() == 0:
			// Add the first non-empty path element unchanged.
		case isSlash(lastChar):
			// If the path ends in a slash, strip any leading slashes from the next
			// path element to avoid creating a UNC path (any path starting with "\\")
			// from non-UNC elements.
			//
			// The correct behavior for Join when the first element is an incomplete UNC
			// path (for example, "\\") is underspecified. We currently join subsequent
			// elements so Join("\\", "host", "share") produces "\\host\share".
			for len(e) > 0 && isSlash(e[0]) {
				e = e[1:]
			}
			// If the path is \ and the next path element is ??,
			// add an extra .\ to create \.\?? rather than \??\
			// (a Root Local Device path).
			if b.Len() == 1 && pathHasPrefixFold(e, "??") {
				b.WriteString(`.\`)
			}
		case lastChar == DriverSeparator:
			// If the path ends in a colon, keep the path relative to the current directory
			// on a drive and don't add a separator. Preserve leading slashes in the next
			// path element, which may make the path absolute.
			//
			// 	Join(`C:`, `f`) = `C:f`
			//	Join(`C:`, `\f`) = `C:\f`
		default:
			// In all other cases, add a separator between elements.
			b.WriteByte(DirSeparator)
			lastChar = DirSeparator
		}
		if len(e) > 0 {
			b.WriteString(e)
			lastChar = e[len(e)-1]
		}
	}
	if b.Len() == 0 {
		return ""
	}

	return Clean(b.String())
}

func isSlash(c uint8) bool {
	return c == DirSeparator || c == '/'
}

// pathHasPrefixFold tests whether the path s begins with prefix,
// ignoring case and treating all path separators as equivalent.
// If s is longer than prefix, then s[len(prefix)] must be a path separator.
// nolint: varnamelen
func pathHasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if isSlash(prefix[i]) {
			if !isSlash(s[i]) {
				return false
			}
		} else if toUpper(prefix[i]) != toUpper(s[i]) {
			return false
		}
	}
	if len(s) > len(prefix) && !isSlash(s[len(prefix)]) {
		return false
	}

	return true
}
func toUpper(c byte) byte {
	if 'a' <= c && c <= 'z' {
		return c - ('a' - 'A')
	}

	return c
}

// volumeNameLen returns length of the leading volume name on Windows.
// It returns 0 elsewhere.
//
// See:
// https://learn.microsoft.com/en-us/dotnet/standard/io/file-path-formats
// https://googleprojectzero.blogspot.com/2016/02/the-definitive-guide-on-win32-to-nt.html
// nolint: mnd
func volumeNameLen(path string) int {
	switch {
	case len(path) >= 2 && path[1] == DriverSeparator:
		// Path starts with a drive letter.
		//
		// Not all Windows functions necessarily enforce the requirement that
		// drive letters be in the set A-Z, and we don't try to here.
		//
		// We don't handle the case of a path starting with a non-ASCII character,
		// in which case the "drive letter" might be multiple bytes long.
		return 2 //nolint: mnd

	case len(path) == 0 || !isSlash(path[0]):
		// Path does not have a volume component.
		return 0

	case pathHasPrefixFold(path, `\\.\UNC`):
		// We're going to treat the UNC host and share as part of the volume
		// prefix for historical reasons, but this isn't really principled;
		// Windows's own GetFullPathName will happily remove the first
		// component of the path in this space, converting
		// \\.\unc\a\b\..\c into \\.\unc\a\c.
		return uncLen(path, len(`\\.\UNC\`))

	case pathHasPrefixFold(path, `\\.`) ||
		pathHasPrefixFold(path, `\\?`) || pathHasPrefixFold(path, `\??`):
		// Path starts with \\.\, and is a Local Device path; or
		// path starts with \\?\ or \??\ and is a Root Local Device path.
		//
		// We treat the next component after the \\.\ prefix as
		// part of the volume name, which means Clean(`\\?\c:\`)
		// won't remove the trailing \. (See #64028.)
		if len(path) == 3 {
			return 3 // exactly \\.
		}
		_, rest, ok := cutPath(path[4:])
		if !ok {
			return len(path)
		}

		return len(path) - len(rest) - 1

	case len(path) >= 2 && isSlash(path[1]):
		// Path starts with \\, and is a UNC path.
		return uncLen(path, 2)
	}

	return 0
}

// postClean adjusts the results of Clean to avoid turning a relative path
// into an absolute or rooted one.
func postClean(out *lazybuf) {
	if out.volLen != 0 || out.buf == nil {
		return
	}
	// If a ListSeparator appears in the path element at the start of a path,
	// insert a .\ at the beginning to avoid converting relative paths
	// like a/../c: into c:.
	for _, c := range out.buf {
		if IsPathSeparator(c) {
			break
		}
		if c == DriverSeparator {
			out.prepend('.', DirSeparator)
			return
		}
	}
	// If a path begins with \??\, insert a \. at the beginning
	// to avoid converting paths like \a\..\??\c:\x into \??\c:\x
	// (equivalent to c:\x).
	if len(out.buf) >= 3 && IsPathSeparator(out.buf[0]) && out.buf[1] == '?' && out.buf[2] == '?' {
		out.prepend(DirSeparator, '.')
	}
}

// uncLen returns the length of the volume prefix of a UNC path.
// prefixLen is the prefix prior to the start of the UNC host;
// for example, for "//host/share", the prefixLen is len("//")==2.
// nolint: mnd
func uncLen(path string, prefixLen int) int {
	count := 0
	for i := prefixLen; i < len(path); i++ {
		if isSlash(path[i]) {
			count++
			if count == 2 {
				return i
			}
		}
	}

	return len(path)
}

// cutPath slices path around the first path separator.
// nolint: nonamedreturns
func cutPath(path string) (before, after string, found bool) {
	for i := range path {
		if isSlash(path[i]) {
			return path[:i], path[i+1:], true
		}
	}

	return path, "", false
}

// IsAbs reports whether the path is absolute.
func IsAbs(path string) (b bool) {
	l := volumeNameLen(path)
	if l == 0 {
		return false
	}

	// If the volume name starts with a double slash, this is an absolute path.
	if isSlash(path[0]) && isSlash(path[1]) {
		return true
	}
	path = path[l:]
	if path == "" {
		return false
	}

	return isSlash(path[0])
}

// A lazybuf is a lazily constructed path buffer.
// It supports append, reading previously appended bytes,
// and retrieving the final string. It does not allocate a buffer
// to hold the output until that output diverges from s.
type lazybuf struct {
	path       string
	buf        []byte
	w          int
	volAndPath string
	volLen     int
}

func (b *lazybuf) index(i int) byte {
	if b.buf != nil {
		return b.buf[i]
	}
	return b.path[i]
}

func (b *lazybuf) append(c byte) {
	if b.buf == nil {
		if b.w < len(b.path) && b.path[b.w] == c {
			b.w++
			return
		}
		b.buf = make([]byte, len(b.path))
		copy(b.buf, b.path[:b.w])
	}
	b.buf[b.w] = c
	b.w++
}

func (b *lazybuf) prepend(prefix ...byte) {
	b.buf = append(prefix, b.buf...)
	b.w += len(prefix)
}

func (b *lazybuf) string() string {
	if b.buf == nil {
		return b.volAndPath[:b.volLen+b.w]
	}
	return b.volAndPath[:b.volLen] + string(b.buf[:b.w])
}

// Clean returns the shortest path name equivalent to path
// by purely lexical processing. It applies the following rules
// iteratively until no further processing can be done:
//
//  1. Replace multiple DirSeparator elements with a single one.
//  2. Eliminate each . path name element (the current directory).
//  3. Eliminate each inner .. path name element (the parent directory)
//     along with the non-.. element that precedes it.
//  4. Eliminate .. elements that begin a rooted path:
//     that is, replace "/.." by "/" at the beginning of a path,
//     assuming DirSeparator is '/'.
//
// The returned path ends in a slash only if it represents a root directory,
// such as "/" on Unix or `C:\` on Windows.
//
// Finally, any occurrences of slash are replaced by DirSeparator.
//
// If the result of this process is an empty string, Clean
// returns the string ".".
//
// On Windows, Clean does not modify the volume name other than to replace
// occurrences of "/" with `\`.
// For example, Clean("//host/share/../x") returns `\\host\share\x`.
//
// See also Rob Pike, “Lexical File Names in Plan 9 or
// Getting Dot-Dot Right,”
// https://9p.io/sys/doc/lexnames.html
func Clean(path string) string {
	originalPath := path
	volLen := volumeNameLen(path)
	path = path[volLen:]
	if path == "" {
		if volLen > 1 && IsPathSeparator(originalPath[0]) && IsPathSeparator(originalPath[1]) {
			// should be UNC
			return FromSlash(originalPath)
		}
		return originalPath + "."
	}
	rooted := IsPathSeparator(path[0])

	// Invariants:
	//	reading from path; r is index of next byte to process.
	//	writing to buf; w is index of next byte to write.
	//	dotdot is index in buf where .. must stop, either because
	//		it is the leading slash or it is a leading ../../.. prefix.
	n := len(path)
	out := lazybuf{path: path, volAndPath: originalPath, volLen: volLen}
	r, dotdot := 0, 0
	if rooted {
		out.append(DirSeparator)
		r, dotdot = 1, 1
	}

	for r < n {
		switch {
		case IsPathSeparator(path[r]):
			// empty path element
			r++
		case path[r] == '.' && (r+1 == n || IsPathSeparator(path[r+1])):
			// . element
			r++
		case path[r] == '.' && path[r+1] == '.' && (r+2 == n || IsPathSeparator(path[r+2])):
			// .. element: remove to last separator
			r += 2
			switch {
			case out.w > dotdot:
				// can backtrack
				out.w--
				for out.w > dotdot && !IsPathSeparator(out.index(out.w)) {
					out.w--
				}
			case !rooted:
				// cannot backtrack, but not rooted, so append .. element.
				if out.w > 0 {
					out.append(DirSeparator)
				}
				out.append('.')
				out.append('.')
				dotdot = out.w
			}
		default:
			// real path element.
			// add slash if needed
			if rooted && out.w != 1 || !rooted && out.w != 0 {
				out.append(DirSeparator)
			}
			// copy element
			for ; r < n && !IsPathSeparator(path[r]); r++ {
				out.append(path[r])
			}
		}
	}

	// Turn empty string into "."
	if out.w == 0 {
		out.append('.')
	}

	postClean(&out) // avoid creating absolute paths on Windows
	return FromSlash(out.string())
}

// ToSlash returns the result of replacing each separator character
// in path with a slash ('/') character. Multiple separators are
// replaced by multiple slashes.
func ToSlash(path string) string {
	return strings.ReplaceAll(path, string(DirSeparator), "/")
}

// FromSlash returns the result of replacing each slash ('/') character
// in path with a separator character. Multiple slashes are replaced
// by multiple separators.
func FromSlash(path string) string {
	return strings.ReplaceAll(path, "/", string(DirSeparator))
}

// SplitList splits a list of paths joined by the OS-specific ListSeparator,
// usually found in PATH or GOPATH environment variables.
// Unlike strings.Split, SplitList returns an empty slice when passed an empty
// string.
func SplitList(path string) []string {
	// The same implementation is used in LookPath in os/exec;
	// consider changing os/exec when changing this.

	if path == "" {
		return []string{}
	}

	// Split path, respecting but preserving quotes.
	list := make([]string, 0)
	start := 0
	quo := false
	for i := 0; i < len(path); i++ {
		switch c := path[i]; {
		case c == '"':
			quo = !quo
		case c == ListSeparator && !quo:
			list = append(list, path[start:i])
			start = i + 1
		}
	}
	list = append(list, path[start:])

	// Remove quotes.
	for i, s := range list {
		list[i] = strings.ReplaceAll(s, `"`, ``)
	}

	return list
}

// Split splits path immediately following the final DirSeparator,
// separating it into a directory and file name component.
// If there is no DirSeparator in path, Split returns an empty dir
// and file set to path.
// The returned values have the property that path = dir+file.
func Split(path string) (dir, file string) {
	vol := VolumeName(path)
	i := len(path) - 1
	for i >= len(vol) && !IsPathSeparator(path[i]) {
		i--
	}
	return path[:i+1], path[i+1:]
}

// Ext returns the file name extension used by path.
// The extension is the suffix beginning at the final dot
// in the final element of path; it is empty if there is
// no dot.
func Ext(path string) string {
	for i := len(path) - 1; i >= 0 && !IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

// Rel returns a relative path that is lexically equivalent to targpath when
// joined to basepath with an intervening separator. That is,
// Join(basepath, Rel(basepath, targpath)) is equivalent to targpath itself.
// On success, the returned path will always be relative to basepath,
// even if basepath and targpath share no elements.
// An error is returned if targpath can't be made relative to basepath or if
// knowing the current working directory would be necessary to compute it.
// Rel calls Clean on the result.
func Rel(basepath, targpath string) (string, error) {
	baseVol := VolumeName(basepath)
	targVol := VolumeName(targpath)
	base := Clean(basepath)
	targ := Clean(targpath)
	if sameWord(targ, base) {
		return ".", nil
	}
	base = base[len(baseVol):]
	targ = targ[len(targVol):]
	if base == "." {
		base = ""
	} else if base == "" && volumeNameLen(baseVol) > 2 /* isUNC */ {
		// Treat any targetpath matching `\\host\share` basepath as absolute path.
		base = string(DirSeparator)
	}

	// Can't use IsAbs - `\a` and `a` are both relative in Windows.
	baseSlashed := len(base) > 0 && base[0] == DirSeparator
	targSlashed := len(targ) > 0 && targ[0] == DirSeparator
	if baseSlashed != targSlashed || !sameWord(baseVol, targVol) {
		return "", errors.New("Rel: can't make " + targpath + " relative to " + basepath)
	}
	// Position base[b0:bi] and targ[t0:ti] at the first differing elements.
	bl := len(base)
	tl := len(targ)
	var b0, bi, t0, ti int
	for {
		for bi < bl && base[bi] != DirSeparator {
			bi++
		}
		for ti < tl && targ[ti] != DirSeparator {
			ti++
		}
		if !sameWord(targ[t0:ti], base[b0:bi]) {
			break
		}
		if bi < bl {
			bi++
		}
		if ti < tl {
			ti++
		}
		b0 = bi
		t0 = ti
	}
	if base[b0:bi] == ".." {
		return "", errors.New("Rel: can't make " + targpath + " relative to " + basepath)
	}
	if b0 != bl {
		// Base elements left. Must go up before going down.
		seps := strings.Count(base[b0:bl], string(DirSeparator))
		size := 2 + seps*3
		if tl != t0 {
			size += 1 + tl - t0
		}
		buf := make([]byte, size)
		n := copy(buf, "..")
		for i := 0; i < seps; i++ {
			buf[n] = DirSeparator
			copy(buf[n+1:], "..")
			n += 3
		}
		if t0 != tl {
			buf[n] = DirSeparator
			copy(buf[n+1:], targ[t0:])
		}
		return string(buf), nil
	}

	return targ[t0:], nil
}

// Base returns the last element of path.
// Trailing path separators are removed before extracting the last element.
// If the path is empty, Base returns ".".
// If the path consists entirely of separators, Base returns a single separator.
func Base(path string) string {
	if path == "" {
		return "."
	}
	// Strip trailing slashes.
	for len(path) > 0 && IsPathSeparator(path[len(path)-1]) {
		path = path[0 : len(path)-1]
	}
	// Throw away volume name
	path = path[len(VolumeName(path)):]
	// Find the last element
	i := len(path) - 1
	for i >= 0 && !IsPathSeparator(path[i]) {
		i--
	}
	if i >= 0 {
		path = path[i+1:]
	}
	// If empty now, it had only slashes.
	if path == "" {
		return string(DirSeparator)
	}

	return path
}

// Dir returns all but the last element of path, typically the path's directory.
// After dropping the final element, Dir calls Clean on the path and trailing
// slashes are removed.
// If the path is empty, Dir returns ".".
// If the path consists entirely of separators, Dir returns a single separator.
// The returned path does not end in a separator unless it is the root directory.
func Dir(path string) string {
	vol := VolumeName(path)
	i := len(path) - 1
	for i >= len(vol) && !IsPathSeparator(path[i]) {
		i--
	}
	dir := Clean(path[len(vol) : i+1])
	if dir == "." && len(vol) > 2 {
		// must be UNC
		return vol
	}

	return vol + dir
}

// VolumeName returns leading volume name.
// Given "C:\foo\bar" it returns "C:" on Windows.
// Given "\\host\share\foo" it returns "\\host\share".
// On other platforms it returns "".
func VolumeName(path string) string {
	return FromSlash(path[:volumeNameLen(path)])
}

func sameWord(a, b string) bool {
	return strings.EqualFold(a, b)
}

// IsPathSeparator reports whether c is a directory separator character.
func IsPathSeparator(c uint8) bool {
	// NOTE: Windows accepts / as path separator.
	return c == DirSeparator || c == '/'
}
