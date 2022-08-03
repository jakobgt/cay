package cay

import (
	"testing"
	"unsafe"

	"github.com/jakobgt/cay/z"
)

func Benchmark_string_len(b *testing.B) {
	keys := RandomKeys(_32K)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entry := n & (_32K - 1)
		foo := len(keys[entry])
		if foo < 0 {
			panic("did not find element")
		}
	}
}

func Benchmark_string_len_reverse(b *testing.B) {
	keys := RandomKeys(_32K)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entry := (b.N - n) & (_32K - 1)
		foo := len(keys[entry])
		if foo+1 < 0 {
			panic("did not find element")
		}
	}
}

func Benchmark_string_struct_len_reverse(b *testing.B) {
	keys := RandomKeys(_32K)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entry := (b.N - n) & (_32K - 1)
		s := (*z.StringStruct)(noescape(unsafe.Pointer(&keys[entry])))
		if s.Len+1 < 0 {
			panic("did not find element")
		}
	}
}
