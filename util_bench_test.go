package cay

// In this function I persist my exploring of the Go language and
// its performance.

import (
	"math/rand"
	"runtime"
	"testing"
	"unsafe"

	"github.com/jakobgt/cay/z"
	"github.com/stretchr/testify/require"
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

const _pageSize = 4 * 1024

type pageSizedEntry struct {
	data [_pageSize]byte
}

func Benchmark_function_call_tlb_miss_rate(b *testing.B) {
	sliceSize := 1 << 20
	sliceMask := sliceSize - 1
	// We make a 1mio array, where each entry is 4KB.
	dataSlice := make([]pageSizedEntry, sliceSize)
	// We populate it with random data
	r := rand.New(rand.NewSource(99))
	tempSlice := make([]byte, _pageSize)
	for i := range dataSlice {
		_, err := r.Read(tempSlice)
		require.NoError(b, err)

		// We copy over the data
		dataSlice[i].data = *(*[_pageSize]byte)(tempSlice)
	}

	// Key generation
	keys := make([]int, sliceSize)
	for i := range keys {
		keys[i] = r.Intn(sliceSize)
	}

	b.ResetTimer()
	b.Run("no function call", func(b *testing.B) {
		sum := 0
		for i := 0; i < b.N; i++ {
			k := keys[i&sliceMask]
			d := &dataSlice[k]
			com := d.data[0] + d.data[1024]
			sum += int(com)
		}
		runtime.KeepAlive(sum)
	})
}
