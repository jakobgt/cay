package cay

import (
	"fmt"
	"unsafe"
)

//go:noescape
func __CompareNMask(buf, mask unsafe.Pointer) (ret uint16)

//go:noescape
//go:nosplit
func __HighestBitMask(buf, unused, result unsafe.Pointer)

//go:nosplit
func CompareNMask(ctrlGroup [_slotsPerGroup]byte, mask byte) uint16 {
	return __CompareNMask(unsafe.Pointer(&ctrlGroup), unsafe.Pointer(uintptr(mask)))
}

//go:nosplit
func HighestBitMask(ctrlGroup [_slotsPerGroup]byte) uint16 {
	var (
		// We don't want ctrlGroup to escape to the heap, but if it doesn't then the code panics.
		// Probably something with alignment....
		p1     = unsafe.Pointer(&ctrlGroup)
		unused = unsafe.Pointer(uintptr(0))
		res    uint16
	)
	//fmt.Printf("%v\n", ctrlGroup)
	__HighestBitMask(p1, unused, unsafe.Pointer(&res))

	// I need this if statement to get the ctrlGroup page-aligned (or something)
	// If I remove this statement, then the call to __HighestBitMask will fail.
	// Note for speed reasons, the if should always be false.
	if unused != unsafe.Pointer(uintptr(0)) {
		fmt.Printf("address: %v\n", p1)
	}
	return res
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input.  noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
//go:nosplit
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
