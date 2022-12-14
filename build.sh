#!/bin/sh

set -e

CLANG=${CLANG:-clang}
C2GOASM=${C2GOASM:-c2goasm}

pushd ./_lib > /dev/null 2>&1

$CLANG -S -O3 -mavx2 -masm=intel -mfma -mno-red-zone -mstackrealign -mllvm -inline-threshold=1000 -fno-asynchronous-unwind-tables -fno-exceptions -fno-rtti -c sum_float64.c

popd > /dev/null 2>&1

# c2goasm
# https://github.com/minio/c2goasm
# https://blog.minio.io/c2goasm-c-to-go-assembly-bb723d2f777f
$C2GOASM -a -f _lib/sum_float64.s sum_float64_simd.s
