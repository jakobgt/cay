#include <immintrin.h>

unsigned int CompareNMask(char buf[], char mask) {
//    // TODO: Should we make the buckets 32 entries (with the AVX2 chipset)?
    __m128i a = _mm_load_si128((__m128i*) buf);
    __m128i b = _mm_set1_epi8(mask);
    __m128i c = _mm_cmpeq_epi8(a, b);
    return _mm_movemask_epi8(c);
}

void HighestBitMask(char buf[], char unused, unsigned int* result) {
//    // TODO: Should we make the buckets 32 entries (with the AVX2 chipset)?
    __m128i a = _mm_load_si128((__m128i*) buf);
    int acc = _mm_movemask_epi8(a);
    *result = acc;
}
