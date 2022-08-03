package cay

import (
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

//var emptyGroup = make([]byte, 16)
var emptyGroup = [_slotsPerGroup]byte{}

func Test__CompareNMask(t *testing.T) {
	ttable := []struct {
		name       string
		inputGroup [_slotsPerGroup]byte
		mask       byte
		wantRest   uint16
	}{
		{
			name:       "empty-g-zero-mask-matches-all",
			inputGroup: emptyGroup,
			mask:       byte(0),
			wantRest:   uint16(0xffff),
		},
		{
			name:       "empty-g-1-mask",
			inputGroup: emptyGroup,
			mask:       byte(1),
			wantRest:   uint16(0),
		},
		{
			name: "1grp-is-one-rest-zero-1-mask-matches-first",
			inputGroup: [_slotsPerGroup]byte{
				1, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
			mask:     byte(1),
			wantRest: uint16(0x1), // First bit is set.
		},
		{
			name: "1grp-is-100-rest-zero-1-mask-matches-nothing",
			inputGroup: [_slotsPerGroup]byte{
				100, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
			mask:     byte(1),
			wantRest: uint16(0x0), // Nothing is set.
		},
		{
			name: "2ndlast-is-100-rest-zero-1-mask-matches-one",
			inputGroup: [_slotsPerGroup]byte{
				0, 100, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
			mask:     byte(100),
			wantRest: uint16(0x0002), // Nothing is set.
		},
	}

	for _, tt := range ttable {
		t.Run(tt.name, func(t *testing.T) {
			res := CompareNMask(tt.inputGroup, tt.mask)

			assert.Equal(t, tt.wantRest, res)
		})
	}
}

func Test__CompareNMask_on_heap_allocated(t *testing.T) {
	testData := &struct {
		data [_slotsPerGroup]byte
	}{
		data: emptyGroup,
	}

	res := CompareNMask(testData.data, byte(0))
	assert.Equal(t, uint16(0xffff), res)
}

func Test__HighestBitMask(t *testing.T) {
	ttable := []struct {
		name       string
		inputGroup [_slotsPerGroup]byte
		wantRest   uint16
	}{
		{
			name:       "empty-bucket",
			inputGroup: emptyGroup,
			wantRest:   uint16(0x0),
		},
		{
			name: "all bytes set",
			inputGroup: [_slotsPerGroup]byte{
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff,
			},
			// We should get
			wantRest: uint16(1<<16 - 1),
		},
		{
			name: "1grp-is-one-rest-zero",
			inputGroup: [_slotsPerGroup]byte{
				0xff, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
			},
			wantRest: uint16(0x1), // First bit is set.
		},
		{
			name: "2-buckets-are-matchins",
			inputGroup: [_slotsPerGroup]byte{
				0xff, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0, 0,
				0, 0, 0xff, 0,
			},
			wantRest: uint16(0x1) + uint16(1<<14), // First and 2nd to last bit is set.
		},
	}
	for _, tt := range ttable {
		t.Run(tt.name, func(t *testing.T) {
			res := HighestBitMask(tt.inputGroup)
			assert.Equal(t, tt.wantRest, res)
		})
	}
}

func Test__HighestBitMask_on_heap_allocated(t *testing.T) {
	testStr := &struct {
		data [_slotsPerGroup]byte
	}{
		data: [_slotsPerGroup]byte{
			0xff, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
	}
	res := HighestBitMask(testStr.data)
	assert.Equal(t, uint16(0x1), res)
}

func Benchmark_CompareNMask(b *testing.B) {
	zeroBucket := [16]byte{}
	fullBucket := [_slotsPerGroup]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
	}
	i := byte(0)
	b.Run("0-bucket and 0-byte", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			CompareNMask(zeroBucket, i)
		}
	})

	b.Run("0-bucket and incrementing mask", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			CompareNMask(zeroBucket, byte(n))
		}
	})

	b.Run("full-bucket and incrementing mask", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			CompareNMask(fullBucket, byte(n))
		}
	})
}

func Benchmark___CompareNMask(b *testing.B) {
	zeroBucket := [16]byte{}
	fullBucket := [_slotsPerGroup]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
	}
	rand.Seed(time.Now().UTC().UnixNano())
	randomBucket := [16]byte{}
	for i := range randomBucket {
		randomBucket[i] = byte(rand.Int())
	}
	i := byte(0)

	b.ResetTimer()
	b.Run("0-bucket and 0-byte", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			__CompareNMask(unsafe.Pointer(&zeroBucket), unsafe.Pointer(uintptr(i)))
		}
	})

	b.Run("0-bucket and incrementing mask", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			__CompareNMask(unsafe.Pointer(&zeroBucket), unsafe.Pointer(uintptr(byte(n))))
		}
	})

	b.Run("full-bucket and incrementing mask", func(b *testing.B) {
		randomByte := byte(rand.Int())
		for n := 0; n < b.N; n++ {
			__CompareNMask(unsafe.Pointer(&fullBucket), unsafe.Pointer(uintptr(randomByte)))
		}
	})

	b.Run("random-bucket and random mask", func(b *testing.B) {
		randomByte := byte(rand.Int())
		for n := 0; n < b.N; n++ {
			__CompareNMask(unsafe.Pointer(&randomBucket), unsafe.Pointer(uintptr(randomByte)))
		}
	})
}
