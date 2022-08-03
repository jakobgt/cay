package cay

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	_8    = 1 << 3
	_1K   = 1 << 10
	_32K  = 1 << 15
	_512K = 1 << 19
	_1M   = 1 << 20

	notFoundKey = "abcdefghijklmnopqrstuva123456789"
)

var (
	_allSizes = []int{_8, _1K, _32K, _512K, _1M}
	_allNames = map[int]string{
		_8:    "8",
		_1K:   "1k",
		_32K:  "32k",
		_512K: "512k",
		_1M:   "1m",
	}

	// Initialized this map up front, to avoid it taking up the traces
	_keysMap = func() map[int][]string {
		toRet := make(map[int][]string, len(_allNames))
		for _, val := range _allSizes {
			toRet[val] = RandomKeys(val)
		}
		return toRet
	}()
)

// To try out adding more memory, to reduce GC.
// var (
// 	_ballast = make([]byte, int64(1)<<int64(33)) // 8GB
// )

func id(keys []string) []string {
	return keys
}

func stringCopy(keys []string) []string {
	fKeys := make([]string, len(keys))

	for i, s := range keys {
		fKeys[i] = fmt.Sprint(s)
	}
	return fKeys
}

func stringCopyRandomOrder(keys []string) []string {
	fKeys := make([]string, len(keys))

	for i, s := range keys {
		fKeys[i] = fmt.Sprint(s)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(keys), func(i, j int) { fKeys[i], fKeys[j] = fKeys[j], fKeys[i] })

	return fKeys
}

func Benchmark_read_identical_string_keys(b *testing.B) {
	for _, val := range _allSizes {
		valToUse := val
		b.Run(_allNames[val], func(b *testing.B) {
			compareCayAndBuiltin(b, _keysMap[valToUse], id, true)
		})
	}
}

func Benchmark_read_not_found_dynamic(b *testing.B) {
	for _, val := range _allSizes {
		valToUse := val
		b.Run(_allNames[val], func(b *testing.B) {
			compareCayAndBuiltin(b, _keysMap[valToUse], func(keys []string) []string {
				return RandomKeysWithSuffix(len(keys), "1")
			}, false)
		})
	}
}

func Benchmark_read_not_found_static_key(b *testing.B) {
	for _, val := range _allSizes {
		valToUse := val
		b.Run(_allNames[val], func(b *testing.B) {
			compareCayAndBuiltin(b, _keysMap[valToUse], func(keys []string) []string {
				return []string{notFoundKey}
			}, false)
		})
	}
}

func Benchmark_read_fresh_string_keys(b *testing.B) {
	compareCayAndBuiltin(b, _keysMap[_32K], stringCopy, true)
}

func Benchmark_read_fresh_string_random_order_keys(b *testing.B) {
	compareCayAndBuiltin(b, _keysMap[_32K], stringCopyRandomOrder, true)
}

func compareCayAndBuiltin(b *testing.B, keys []string, m func(keys []string) []string, present bool) {
	b.Helper()
	caymap := Simdmap(keys)
	val := bmap(keys)

	// We now update the set of keys to lookup via
	keys = m(keys)
	// We assert that keys is a 2 exponential
	require.Equal(b, 0, len(keys)&(len(keys)-1))
	// Such that we can set the mask to length minus one.
	lenMask := len(keys) - 1
	var lenKept int
	debug.SetGCPercent(-1)
	b.ResetTimer()

	b.Run("caymap", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			entry := n & lenMask
			v, ok := caymap.Get(keys[entry])
			lenKept = len(v)
			if present && !ok {
				panic("did not find element")
			} else if !present && ok {
				panic("did find element")
			}
		}
	})

	b.Run("builtin", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			entry := n & lenMask
			v, ok := val[keys[entry]]
			lenKept = len(v)
			if present && !ok {
				panic("did not find element")
			} else if !present && ok {
				panic("did find element")
			}
		}
	})

	runtime.KeepAlive(lenKept)
}

// func Benchmark_builtin_map_insert(b *testing.B) {
// 	keys := RandomKeys(b.N)
// 	b.ResetTimer()
// 	bmap(keys)
// }

// func Benchmark_simdmap_insert(b *testing.B) {
// 	keys := RandomKeys(b.N)
// 	val := NewMap(b.N)
// 	for n := 0; n < b.N; n++ {
// 		val.Insert(keys[n], []byte("data"))
// 	}
// }

// func Benchmark_bit_iterations_for_loop(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		idx := uint16(1 << 10)
// 		for i := uint16(0); i < 16; i++ {
// 			// If there is no 1 bit at this position, then we skip
// 			if (idx & 1) == 0 {
// 				idx = idx >> 1
// 				continue
// 			}
// 			idx = idx >> 1
// 		}
// 	}
// }

// var foo uint16

// func Benchmark_bit_iterations_trailing_zeros(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		idx := uint16(1 << 10)
// 		for idx != 0 {
// 			t := idx & -idx
// 			bits.TrailingZeros16(idx)
// 			//callback(k*64 + r)
// 			idx ^= t
// 		}
// 	}
// }

// func Benchmark_bit_xoring(b *testing.B) {
// 	idx := uint16(1 << 10)
// 	for n := 0; n < b.N; n++ {
// 		t := idx & -idx
// 		idx ^= t
// 	}
// }

// var fooo int64

// func Benchmark_uint64_to_int64(b *testing.B) {
// 	t := int64(0)
// 	for n := 0; n < b.N; n++ {
// 		idx := uint64(n)
// 		t = int64(idx)
// 	}
// 	fooo = t
// }
