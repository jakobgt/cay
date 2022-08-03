package cay

import (
	"math/bits"
	"unsafe"

	"github.com/jakobgt/cay/z"
)

const (
	PtrSize = 4 << (^uintptr(0) >> 63)
)

// Get returns the entry with the given key and a bool if the key was found.
func (m *Map[V]) Get(key string) (V, bool) {
	p, ok := m.findGet(key)
	if !ok {
		var noop V
		return noop, false
	}
	return *(*V)(p), true
}

// findGet returns a pointer to value if present in the map and a boolean that represent if
// the key was found
func (m *Map[V]) findGet(key string) (*V, bool) {
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
		bucket := &m.buckets[cGroup]
		ctrl := bucket.controls
		grpCtrlPointer := noescape(unsafe.Pointer(&ctrl))
		// It might seem that __CompareNMask takes a long time, but that is because it hits a
		// TLB miss.
		// I wonder whether this call could mess up the TLB?
		idx := __CompareNMask(grpCtrlPointer, unsafe.Pointer(hash>>57))
		bEntries := &bucket.entries
		// bKey := bEntries[0].key
		// bKey = bEntries[4].key
		// bKey = bEntries[8].key
		// bKey = bEntries[12].key

		// IDX has 1 in its bit representation for every match, and so we iterate each of these
		// positions.
		// This is a quick way to do it, and I don't quite understand it, but it is fast
		// (5x over the naive for loop)
		// More info in https://lemire.me/blog/2018/02/21/iterating-over-set-bits-quickly/
		for idx != 0 {
			i := bits.TrailingZeros16(idx)
			// Somewhat annoying, but here we get a bounds check on i and bEntries, even though
			// i is always lower than len(bEntries). Maybe we can just look it up directly
			entry := &bEntries[i]
			bKey := entry.key
			//ctrl = bucket.controls
			// With the mask of m and idx overlapping there is a potential candidate at this
			// pos that could be the key
			//			absPos := cGroup*16 + uint64(i)
			// These two lines are equal, I'm just testing what the extra index bound chek means in
			// terms of performance.
			ekeyP := (*z.StringStruct)(unsafe.Pointer(&bKey))
			//ekeyP := (*z.StringStruct)(add(unsafe.Pointer(&bucket.entries), uintptr(i)*entrySize+keyOffset))

			// This comparison might seem slow, but it is because the ekeyP.Len hits a TLB miss ~16% of the
			// time. Unsure why (the page that ekeyP is part of should have been loaded with __CompareNMask).
			// Furthermore this causes 77% of the total cache misses (in just this function - 1818 samples, whereas the
			// the __CompareNMask accounts for 1752 cache misses -excluding TLB misses. )
			if keyP.Len != ekeyP.Len {
				idx ^= idx & -idx
				continue

			}
			// We force i to be less than 16 to avoid the index out of bound check below.
			// This causes a ton of TLB misses.
			if keyP.Str == ekeyP.Str {
				return &entry.value, true
			}

			if z.Memequal(ekeyP.Str, keyP.Str, uintptr(keyP.Len)) {
				// TODO, we need to ensure that i is less than 16 to avoid a out of bound check
				//				fmt.Printf("keyP and ekeyP do not point to the same str: &keyP: %x and &ekeyP: %x\n", keyP.Str, ekeyP.Str)
				return &entry.value, true
			}
			idx ^= idx & -idx
			// if idx != 0 {
			// 	fmt.Println("Found multiple hashes that matched")
			// }
		}

		// Check whether any slot in the bucket is empty. If so we've found our bucket
		if !bucket.full {
			break
		}
	}
	// We did not find the values
	return (*V)(nil), false
}
