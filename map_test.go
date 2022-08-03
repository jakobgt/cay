package cay

import (
	"fmt"
	"strconv"
	"testing"
	"unsafe"

	"github.com/jakobgt/cay/z"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test__BMap(t *testing.T) {
	m := map[string]struct{}{}

	m["jakob"] = struct{}{}

	_, ok := m[fmt.Sprintf("jakob")]
	assert.True(t, ok)
}

func Test__Map(t *testing.T) {
	m := NewMap(128)

	_, ok := m.Get("foobar")
	assert.False(t, ok)
}

func Test__FirstBucketIsPageAligned(t *testing.T) {
	m := NewMap(128)

	_, ok := m.Get("foobar")
	assert.False(t, ok)

	fAddr := uintptr(unsafe.Pointer(&m.buckets[0]))
	assert.Equal(t, uintptr(0), fAddr%4096, "Bucket[0] with address: %x is not page-aligned", fAddr)
}

func Test__FirstKeyEntryIsSameAddressAsKeys(t *testing.T) {
	m := NewMap(128)

	_, ok := m.Get("foobar")
	assert.False(t, ok)

	ksAddr := uintptr(unsafe.Pointer(&m.buckets[0].entries))
	k0Addr := uintptr(unsafe.Pointer(&m.buckets[0].entries[0].key))
	assert.Equal(t, ksAddr, k0Addr)
}

type aStruct struct {
	foobar [16]byte
}

func Test__SizeOfByteSlice(t *testing.T) {
	//t.Skip("Just for fun and sizing structs in Golang.")
	fmt.Printf("[]byte{}: %d\n", unsafe.Sizeof([]byte{}))
	fmt.Printf("*[]byte{}: %d\n", unsafe.Sizeof(&[]byte{}))
	fmt.Printf("[2]byte{}: %d\n", unsafe.Sizeof([2]byte{}))
	fmt.Printf("[16][]byte{}: %d\n", unsafe.Sizeof([16][]byte{}))
	fmt.Printf("[16]string: %d\n", unsafe.Sizeof([16]string{}))
	fmt.Printf("\"\": %d\n", unsafe.Sizeof(""))
	fmt.Printf("\"foobar\": %d\n", unsafe.Sizeof("foobar"))
	fmt.Printf("aStruct: %d\n", unsafe.Sizeof(aStruct{}))
	fmt.Printf("bucket: %d\n", unsafe.Sizeof(bucket{}))
	fmt.Printf("[]byte(nil): %d\n", unsafe.Sizeof([]byte((nil))))
	fmt.Printf("entry{}: %d\n", unsafe.Sizeof(entry{}))
	//assert.Fail(t, "foobar")
}

func Test__lowestPowerOfTwo(t *testing.T) {
	ttable := []struct {
		input       int
		wantLogSize uint8
	}{
		{
			input:       0,
			wantLogSize: uint8(0),
		},
		{
			input:       5,
			wantLogSize: uint8(0),
		},
		{
			input:       16,
			wantLogSize: uint8(1),
		},
		{
			input:       17,
			wantLogSize: uint8(1),
		},
		{
			input:       32,
			wantLogSize: uint8(2),
		},
		{
			input:       33,
			wantLogSize: uint8(2),
		},
		{
			input:       127,
			wantLogSize: uint8(3),
		},
	}

	for _, tt := range ttable {
		t.Run(strconv.Itoa(tt.input), func(t *testing.T) {
			pos := logSizeOfBuckets(tt.input)
			assert.Equal(t, tt.wantLogSize, pos, "the number of buckets does not match")
		})
	}
}

func Test__bucketMask(t *testing.T) {
	ttable := []struct {
		name     string
		mapSize  int
		wantMask uintptr
	}{
		{
			name:     "size 15 - and it should give a single bucket",
			mapSize:  15,
			wantMask: uintptr(0),
		},
		{
			name:     "size 128 - and it should the mask for 8 buckets (i.e., the value 7)",
			mapSize:  127,
			wantMask: uintptr(8 - 1),
		},
	}

	for _, tt := range ttable {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMap(tt.mapSize)
			bm := bucketMask(m.logSize)
			assert.Equal(t, tt.wantMask, bm)
		})
	}
}

func Test__hashKey(t *testing.T) {
	ttable := []struct {
		name       string
		mapSize    int
		wantBucket uintptr
	}{
		{
			name:       "empty string in a single bucket",
			mapSize:    5,
			wantBucket: uintptr(0),
		},
	}

	for _, tt := range ttable {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMap(tt.mapSize)
			_, bucket := m.hashKey(tt.name)

			assert.Equal(t, tt.wantBucket, bucket)
		})
	}
}

func Test__Map_can_insert_and_retrieve_a_value_from_one_bucket(t *testing.T) {
	m := NewMap(15)

	key := "foobar"
	value := []byte("data")

	m.Insert(key, value)
	val, ok := m.Get(key)
	require.True(t, ok)
	assert.Equal(t, value, val)
}

func Test__Map_can_insert_and_retrieve_a_value(t *testing.T) {
	m := NewMap(16)

	key := "foobar"
	value := []byte("data")
	pKey := (*z.StringStruct)(unsafe.Pointer(&key))
	fmt.Printf("&pKey: %x", pKey.Str)
	m.Insert(key, value)
	val, ok := m.Get(key)
	require.True(t, ok)
	assert.Equal(t, value, val)
}

func Test__Map_can_insert_two_values_in_same_bucket(t *testing.T) {
	m := NewMap(16)

	m.Insert("foobar", []byte("data"))
	m.Insert("bingo", []byte("data"))
}

func Test__Map_insert_and_retrieve_2_values(t *testing.T) {
	size := 2
	m := NewMap(size)

	insertEntries(m, size)
	verifyEntries(t, m, size)
}

func Test__Map_insert_and_retrieve_8_values(t *testing.T) {
	size := 6
	m := NewMap(size)

	insertEntries(m, size)
	verifyEntries(t, m, size)
}

func Test__Map_insert_and_retrieve_15_values(t *testing.T) {
	size := 15
	m := NewMap(size)

	insertEntries(m, size)
	verifyEntries(t, m, size)
}

func Test__Map_insert_and_retrieve_16_values(t *testing.T) {
	size := 16
	m := NewMap(size)

	insertEntries(m, size)
	verifyEntries(t, m, size)
}

func Test__Map_can_insert_16_entries_in_two_groups(t *testing.T) {
	m := NewMap(16)
	insertEntries(m, 16)

	assert.Equal(t, uint64(16), m.usedSpace)
}

func Test__Map_can_insert_32_entries_in_two_groups(t *testing.T) {
	size := uint64(32)
	m := NewMap(int(size))
	insertEntries(m, int(size))

	assert.Equal(t, size, m.usedSpace)
}

func Test__Map_insert_and_retrieve_on_big_map(t *testing.T) {
	m := NewMap(64)

	m.Insert("42", []byte("42"))
	val, ok := m.Get("42")
	assert.True(t, ok)
	assert.Equal(t, []byte("42"), val)
}

func Test__Map_insert_and_retrieve_32_values(t *testing.T) {
	entries := 32
	m := NewMap(512)

	insertEntries(m, int(entries))

	assert.Equal(t, uint64(entries), m.usedSpace)
	verifyEntries(t, m, entries)
}

func Test__Map_insert_many_values(t *testing.T) {
	entries := 1 << 5
	m := NewMap(entries)
	insertEntries(m, entries)

	assert.Equal(t, uint64(entries), m.usedSpace, "in non-hex: %d", m.usedSpace)
	verifyEntries(t, m, entries)
	fmt.Println(m.noMatch)
	// all is good.
}

func insertEntries(m *Map, size int) {
	for i := uint64(0); i < uint64(size); i++ {
		m.Insert(strconv.Itoa(int(i)), []byte("data"))
	}
}

func verifyEntries(t *testing.T, m *Map, size int) {
	for i := uint64(0); i < uint64(size); i++ {
		val, ok := m.Get(strconv.Itoa(int(i)))
		assert.True(t, ok)
		assert.Equal(t, []byte("data"), val)
	}
}
