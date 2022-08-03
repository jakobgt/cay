package cay

import (
	"fmt"
	"math/bits"
	"unsafe"

	"github.com/jakobgt/cay/z"
)

const (
	_slotsPerGroup         = 16
	_EmptySlots    byte    = 255
	_notFound      uintptr = 1 << 63
)

//
//type entry struct {
//	key   string
//	value []byte
//}

// TODO Bucket on an entry? That way the
type bucket struct {
	controls [_slotsPerGroup]byte
	keys     [_slotsPerGroup]string
	full     bool
	values   [_slotsPerGroup][]byte
}

type Map struct {
	//// We divide these into keys and values for locality
	buckets []bucket
	// These masks are for quickly locating the slot and group from a hash
	logSize uint8 // log_2 of # of groups (can hold up to loadFactor * 2^B items)

	// noOfGroups is a convenience field containing the number of groups
	noOfGroups uint64

	usedSpace uint64

	noMatch uint64
}

func NewMap(size int) *Map {
	logSize := logSizeOfBuckets(size)

	m := &Map{
		logSize: logSize,

		//		buckets: make([]bucket, powerOfTwoSize/_slotsPerGroup),
		// The logSize of buckets is used to derive the number of buckets
		buckets: make([]bucket, 1<<logSize),
	}
	// We need to mark all entries as empty
	for gi := range m.buckets {
		g := &m.buckets[gi]
		g.controls = [_slotsPerGroup]byte{}
		for i := range g.controls {
			g.controls[i] = 255
		}
		//g.values = make([][]byte, _slotsPerGroup)
		//g.keys = make([]string, _slotsPerGroup)
	}
	return m
}

func (m *Map) Insert(key string, value []byte) {
	hash, sGroup := m.hashKey(key)
	// First, we iterate the groups to figure out whether the key is already in any of the groups
	grpF, pos, _ := m.find(key)

	// If we found the key in the map, then we return
	if grpF != _notFound {
		m.buckets[grpF].values[pos] = value
		return
	}

	bMask := bucketMask(m.logSize)

	// At this point we need to find then the first empty or deleted tombstone.
	iterations := 0
	for cGroup := sGroup; iterations < len(m.buckets); iterations, cGroup = iterations+1, (cGroup+1)&bMask {
		grp := &m.buckets[cGroup]

		// Now we just search for the first empty or delete slot
		bitvec := HighestBitMask(grp.controls)
		if bitvec == 0 {
			continue
		}

		emptySlot := bits.TrailingZeros16(bitvec)
		// We want to fill the group from right to left to ease caching and memory prefetching
		//absSlot := cGroup*16 + uint64(emptySlot)
		grp.controls[emptySlot] = byte(hash >> 57)
		grp.keys[emptySlot] = key
		grp.values[emptySlot] = value

		// Make sure to set the full bit.
		// Count ones in the bitvector. If there is less than 2, then this bucket is now full.
		if bits.OnesCount16(bitvec) < 2 {
			grp.full = true
		}

		m.usedSpace++
		return
	}

	panic(fmt.Sprintf("Not enough space in the hashmap for key %s and mask %d (calculated bucket: %d). Size is %d and used space is %d. Tried %d buckets\n",
		key, bMask, sGroup, len(m.buckets), m.usedSpace, sGroup))
}

func (m *Map) Get(key string) ([]byte, bool) {
	grp, _, val := m.find(key)

	if grp == _notFound {
		return nil, false
	}

	return *val, true
}

const (
	PtrSize = 4 << (^uintptr(0) >> 63)
)

// find returns the position of the key in the map. You need to pass in the bucket and mask of the
// key, as returned from hashKey
func (m *Map) find(key string) (group uintptr, slot int, value *[]byte) {
	// Manually inlining the hashKey function as the Go compiler won't
	keyP := (*z.StringStruct)(unsafe.Pointer(&key))

	hash := z.Memhash(keyP.Str, 0, uintptr(keyP.Len))
	// The control mask is the 7 highest bits of the hash.

	// We find the slot that this hash belongs to. We need to use a bit mask as modulo is expensive
	// when the divisor is unknown to the compiler.

	// Then we need to find the bucket, in which the slot belongs
	// And to find the bucket we divide by the size of the bucket. Since this is a power of two
	// and slot is unsigned the compiler automatically changes this to a shift operation.
	// Note that we don't store this value, as it is faster to just compute it from the logSize of
	// the number of buckets.
	//bMask := bucketMask(m.logSize)
	bMask := bucketMask(m.logSize)
	sGroup := hash & bMask // Equal to hash & m.slotMask / 16

	for cGroup := sGroup; cGroup < uintptr(len(m.buckets)); cGroup = (cGroup + 1) & bMask {
		//	for cGroup := sGroup; cGroup < uintptr(len(m.buckets)); cGroup = (cGroup + 1) & bMask {
		grp := &m.buckets[cGroup]
		grpCtrlPointer := unsafe.Pointer(&grp.controls)
		idx := __CompareNMask(grpCtrlPointer, unsafe.Pointer(hash>>57))
		// IDX has 1 in its bit representation for every match, and so we iterate each of these
		// positions.
		// This is a quick way to do it, and I don't quite understand it, but it is fast
		// (5x over the naive for loop)
		// More info in https://lemire.me/blog/2018/02/21/iterating-over-set-bits-quickly/
		for idx != 0 {
			t := idx & -idx
			i := bits.TrailingZeros16(idx)
			idx ^= t
			// With the mask of m and idx overlapping there is a potential candidate at this
			// pos that could be the key
			//			absPos := cGroup*16 + uint64(i)
			// This is slow? Key comparison? And the generated instructions have a JMPQ in it,
			// so it might the compuler
			// It might seem that the bounds on this index lookup is what is causing it
			// to be slow...
			ekeyP := (*z.StringStruct)(unsafe.Pointer(&grp.keys[i]))

			if keyP.Len != ekeyP.Len {
				continue
			}
			// We force i to be less than 16 to avoid the index out of bound check below.
			if keyP == ekeyP || keyP.Str == ekeyP.Str || z.Memequal(ekeyP.Str, keyP.Str, uintptr(keyP.Len)) {
				// TODO, we need to ensure that i is less than 16 to avoid a out of bound check
				return cGroup, i, &grp.values[i]
			}
		}

		// Check whether any slot in the bucket is empty. If so we've found our bucket
		if !grp.full {
			break
		}
	}
	// We did not find the values
	return _notFound, 0, nil
}

func (m *Map) hashKey(key string) (hash, bucket uintptr) {
	keyP := (*z.StringStruct)(unsafe.Pointer(&key))

	hash = z.Memhash(keyP.Str, 0, uintptr(keyP.Len))
	// The control mask is the 7 highest bits of the hash.

	// We find the slot that this hash belongs to. We need to use a bit mask as modulo is expensive
	// when the divisor is unknown to the compiler.

	// Then we need to find the bucket, in which the slot belongs
	// And to find the bucket we divide by the size of the bucket. Since this is a power of two
	// and slot is unsigned the compiler automatically changes this to a shift operation.
	bMask := bucketMask(m.logSize)
	bucket = hash & bMask // Equal to hash & m.slotMask / 16	return
	return
}

// logSizeOfBuckets returns the the number of buckets required to hold the given size.
func logSizeOfBuckets(size int) uint8 {
	buckets := size / 16
	// Min size is 16
	if buckets <= 0 {
		return uint8(0)
	}
	// We round to closest 2 exponent
	pos := bits.LeadingZeros64(uint64(buckets))
	return uint8(64 - uint64(pos))
}

// bucketShift and bucketMask are taken from runtime/map.go

// bucketShift returns 1<<b, optimized for code generation.
func bucketShift(b uint8) uintptr {
	// Masking the shift amount allows overflow checks to be elided.
	return uintptr(1) << (b & (PtrSize*8 - 1))
}

// bucketMask returns 1<<b - 1, optimized for code generation.
func bucketMask(b uint8) uintptr {
	return bucketShift(b) - 1
}
