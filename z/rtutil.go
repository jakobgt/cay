// Copyright (c) 2019 Ewan Chou

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Copied from https://github.com/dgraph-io/ristretto/tree/master/z

package z

import (
	"unsafe"
)

// NanoTime returns the current time in nanoseconds from a monotonic clock.
//go:linkname NanoTime runtime.nanotime
func NanoTime() int64

// CPUTicks is a faster alternative to NanoTime to measure time duration.
//go:linkname CPUTicks runtime.cputicks
func CPUTicks() int64

type StringStruct struct {
	Str unsafe.Pointer
	Len int
}

//go:noescape
//go:linkname Memhash runtime.memhash
func Memhash(p unsafe.Pointer, h, s uintptr) uintptr

// MemHash is the hash function used by go map, it utilizes available hardware instructions(behaves
// as aeshash if aes instruction is available).
// NOTE: The hash seed changes for every process. So, this cannot be used as a persistent hash.
func MemHash(data []byte) uint64 {
	ss := (*StringStruct)(unsafe.Pointer(&data))
	return uint64(Memhash(ss.Str, 0, uintptr(ss.Len)))
}

// MemHashString is the hash function used by go map, it utilizes available hardware instructions
// (behaves as aeshash if aes instruction is available).
// NOTE: The hash seed changes for every process. So, this cannot be used as a persistent hash.
func MemHashString(str string) uint64 {
	ss := (*StringStruct)(unsafe.Pointer(&str))
	return uint64(Memhash(ss.Str, 0, uintptr(ss.Len)))
}

//go:noescape
//go:linkname aeshash64 runtime.aeshash64
func aeshash64(p unsafe.Pointer, h uintptr) uintptr

// MemHashUint64 uses the aeshash given by the golang runtime.
func MemHashUint64(val uint64) uint64 {
	return uint64(aeshash64(unsafe.Pointer(&val), 0))
}

// FastRand is a fast thread local random function.
//go:linkname FastRand runtime.fastrand
func FastRand() uint32

//go:noescape
//go:linkname Memequal runtime.memequal
func Memequal(a, b unsafe.Pointer, size uintptr) bool

//// MemEqualString compares two strings for equality.
//func MemEqualString(a unsafe.Pointer, b *StringStruct) bool {
//	as := (*StringStruct)(a)
//	// Short circuit it, if we have the same key
//	if as.str == b.str {
//		return true
//	}
//
//	if as.len != b.len {
//		return false
//	}
//
//	return memequal(as.str, b.str, uintptr(as.len))
//}
