	.section	__TEXT,__text,regular,pure_instructions
	.build_version macos, 10, 15	sdk_version 10, 15
	.intel_syntax noprefix
	.globl	_CompareNMask           ## -- Begin function CompareNMask
	.p2align	4, 0x90
_CompareNMask:                          ## @CompareNMask
## %bb.0:
	push	rbp
	mov	rbp, rsp
	and	rsp, -8
	vmovd	xmm0, esi
	vpbroadcastb	xmm0, xmm0
	vpcmpeqb	xmm0, xmm0, xmmword ptr [rdi]
	vpmovmskb	eax, xmm0
	mov	rsp, rbp
	pop	rbp
	ret
                                        ## -- End function
	.globl	_HighestBitMask         ## -- Begin function HighestBitMask
	.p2align	4, 0x90
_HighestBitMask:                        ## @HighestBitMask
## %bb.0:
	push	rbp
	mov	rbp, rsp
	and	rsp, -8
	vmovdqa	xmm0, xmmword ptr [rdi]
	vpmovmskb	eax, xmm0
	mov	dword ptr [rdx], eax
	mov	rsp, rbp
	pop	rbp
	ret
                                        ## -- End function

.subsections_via_symbols
