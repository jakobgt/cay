//+build !noasm !appengine
// AUTO-GENERATED BY C2GOASM -- DO NOT EDIT

// func __CompareNMask(buf, mask unsafe.Pointer) (ret uint16)
TEXT ·__CompareNMask(SB), $0-24

	MOVQ buf+0(FP), DI
	MOVQ mask+8(FP), SI

	LONG $0xc66ef9c5             // vmovd    xmm0, esi
	LONG $0x7879e2c4; BYTE $0xc0 // vpbroadcastb    xmm0, xmm0
	LONG $0x0774f9c5             // vpcmpeqb    xmm0, xmm0, oword [rdi]
	LONG $0xc0d7f9c5             // vpmovmskb    eax, xmm0
	// The result has to be 8 byte aligned, so the result is at 16 + FP
	MOVW AX, ret+16(FP)
	RET

TEXT ·__HighestBitMask(SB), $0-24

	MOVQ buf+0(FP), DI
	MOVQ unused+8(FP), SI
	MOVQ result+16(FP), DX

	LONG $0x076ff9c5 // vmovdqa    xmm0, oword [rdi]
	LONG $0xc0d7f9c5 // vpmovmskb    eax, xmm0
	WORD $0x0289     // mov    dword [rdx], eax
	RET