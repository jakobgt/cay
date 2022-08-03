package cay

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
)

const (
	///////// 1 000 000 000
	entries = 1 << 15
	lenMask = entries - 1

	notFoundKey = "abcdefghijklmnopqrstuva123456789"
)

// To try out adding more memory, to reduce GC.
// var (
// 	_ballast = make([]byte, int64(1)<<int64(33)) // 8GB
// )

//var _ = func() int {
//	fmt.Println("sleeping")
//	time.Sleep(time.Second * 15)
//	fmt.Println("starting")
//	return 1
//}()

//func Benchmark_all_together(b *testing.B) {
//	keys := randomKeys(entries)
//	val := bmap(keys)
//	m := Simdmap(keys)
//	b.ResetTimer()
//
//}

var _lenKept int
var _vKept byte

func id(keys []string) []string {
	return keys
}

func stringCopy(keys []string) []string {
 	fKeys := make([]string, entries)

	for i, s := range keys {
		fKeys[i] = fmt.Sprintf(s)
	}
	return fKeys
}

func stringCopyRandomOrder(keys []string) []string {
 	fKeys := make([]string, entries)

	for i, s := range keys {
		fKeys[i] = fmt.Sprintf(s)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(keys), func(i, j int) { fKeys[i], fKeys[j] = fKeys[j], fKeys[i] })

	return fKeys
}

func Benchmark_read_identical_string_keys(b *testing.B) {
	keys := RandomKeys(entries)
	compareCayAndBuiltin(b, keys, id)
}

func Benchmark_read_fresh_string_keys(b *testing.B) {
	keys := RandomKeys(entries)
	compareCayAndBuiltin(b, keys, stringCopy)
}

func Benchmark_read_fresh_string_random_order_keys(b *testing.B) {
	keys := RandomKeys(entries)
	compareCayAndBuiltin(b, keys, stringCopyRandomOrder)
}

func compareCayAndBuiltin(b *testing.B, keys []string, m func(keys []string) []string) {
	b.Helper()
	caymap := Simdmap(keys)
	val := bmap(keys)

	// We now update the set of keys to lookup via
	keys = m(keys)
	var lenKept int
	// To reduce GC impact, we turn it off here.
	runtime.GC()
	debug.SetGCPercent(-1)
	b.ResetTimer()

	b.Run("caymap", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			entry := n & lenMask
			v, ok := caymap.Get(keys[entry])
			lenKept = len(v)
			if !ok {
				panic("did not find element")
			}
		}
	})

	b.Run("builtin", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			entry := n & lenMask
			v, ok := val[keys[entry]]
			lenKept = len(v)
			if !ok {
				panic("did not find element")
			}
		}
	})

	runtime.KeepAlive(lenKept)
}

func Benchmark_builtin_map_get_not_found(b *testing.B) {
	keys := RandomKeys(entries)
	val := bmap(keys)
	notFoundKeys := RandomKeys(entries)
	b.ResetTimer()

	b.Run("static key", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			v, ok := val[notFoundKey]
			if ok {
				fmt.Print(v)
			}
		}
	})

	b.Run("dynamic keys", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			v, ok := val[notFoundKeys[n&lenMask]]
			if ok {
				fmt.Print(v)
			}
		}
	})
}

func Benchmark_simdmap_get_not_found(b *testing.B) {
	keys := RandomKeys(entries)
	m := Simdmap(keys)
	//	full := 0
	// for _, v := range m.buckets {
	// 	if v.full {
	// 		full++
	// 	}
	// }

	// fmt.Printf("Buckets: %d, of which %d are full \n", len(m.buckets), full)

	notFoundKeys := RandomKeysWithSuffix(entries, "1")
	b.ResetTimer()

	b.Run("static key", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_, ok := m.Get(notFoundKey)
			if !ok {
				//Nothing
			}
		}
	})

	b.Run("dynamic keys", func(b *testing.B) {
		//fmt.Println(unsafe.Sizeof(bucket{}))
		for n := 0; n < b.N; n++ {
			v, ok := m.Get(notFoundKeys[n&lenMask])
			if ok {
				fmt.Print(v)
			}
		}
	})
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

func bmap(keys []string) map[string][]byte {
	val := make(map[string][]byte, len(keys))
	for n := 0; n < len(keys); n++ {
		val[keys[n]] = []byte("data")
	}
	return val
}
