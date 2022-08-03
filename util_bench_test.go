package cay

import (
	"testing"
	"unsafe"

	"github.com/jakobgt/cay/z"
)

func Benchmark_string_len(b *testing.B) {
	keys := RandomKeys(entries)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entry := n & lenMask
		foo := len(keys[entry])
		if foo < 0 {
			panic("did not find element")
		}
	}
}

func Benchmark_string_len_reverse(b *testing.B) {
	keys := RandomKeys(entries)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entry := (b.N - n) & lenMask
		foo := len(keys[entry])
		if foo+1 < 0 {
			panic("did not find element")
		}
	}
}

func Benchmark_string_struct_len_reverse(b *testing.B) {
	keys := RandomKeys(entries)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entry := (b.N - n) & lenMask
		s := (*z.StringStruct)(noescape(unsafe.Pointer(&keys[entry])))
		if s.Len+1 < 0 {
			panic("did not find element")
		}
	}
}