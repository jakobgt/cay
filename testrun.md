# Observation log

As this is my first foray into optimizing memory accessed and what not in Go, I try to document my findings here (also to flush my mental cach).

Run date not known
==================

- Not inlining a function can add up to 4ns
- Adding `-cpuprofile cpu.profile ` adds a few ns.
- Using `groupMask(...)` to generate the group mask over (hash & m.Mask / 16) shaves off a ns/op (and the mask generate is dropped by 4.5x from 70ms to 20ms).
  Golang then uses the `MOVZX` operation over the `MOVQ`.
    Code went from
```
   148         70ms       70ms           	sGroup := hash & m.slotMask / 16
                40ms       40ms  129b7db:             MOVQ 0x90(SP), CX                                                    map.go:148
                20ms       20ms  129b7e3:             MOVQ 0x18(CX), DX                                                    map.go:148
                   .          .  129b7e7:             ANDQ AX, DX                                                          map.go:148
                10ms       10ms  129b7ea:             SHRQ $0x4, DX                                                        map.go:148

```
  To:

```
     150         20ms       20ms           	groupMask := bucketMask(m.logSize)
                 20ms       20ms  129b7fb:             MOVQ 0x90(SP), CX                                                    map.go:150
                    .          .  129b803:             MOVZX 0x18(CX), DX                                                   map.go:150

     151            .          .           	sGroup := hash & groupMask // Equal to hash & m.slotMask / 16
                    .          .  129b818:             ANDQ AX, CX
```
- The native hashmap does a lot to avoid bound checks (e.g., for bitmasks it uses uintptr that)
- The max key size for builtin map is 128 (bits/bytes?) before it is not inlined. Similarly for an elem size.
- Letting `__CompareNMask` return an int instead of returning via an argument does not change a lot (maybe 1 ns/op or so.)
- ~0.8% of the SIMD buckets are full (~1k out of 131k bucket) causing ~10ns extra time per op, if we need to search the next bucket, instead
  of stopping after searching the first bucket.
- Having different sized keys result in a drop of 12ns/op (from 100ns to 88ns) just as we avoid comparing keys.
- Wow, the caching nature of an open addressing hashmap is crazy (20ns/op for varying keys with SIMD vs. 150 for builtin)
  The differing factor is that it 7x more to fetch the buckets from memory (specifically the `MOVZX 0(BX)(CX*1), R8 ` operation on `map_faststr.go:191`.
```
~/g/s/c/i/p/l/simdmap (simdmap) $ go test -run=^\$ -benchmem -cpu 1 -bench '.*not_found.*'                                                                         17:35 17/03
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_simdmap_get_not_found/static_key         	85533632	        14.3 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get_not_found/dynamic_keys       	59230143	        19.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/static_key     	87805620	        13.8 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/dynamic_key    	 8325208	       149 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/jakobgt/cay	8.869s
```




Run 2020-11-30
##############

```
 $ go version                                                                                                            20:32 30/11
go version go1.15.2 darwin/amd64
```
Local Mac (all tests are green)

Run (simdmap wins):
```
$ go test -run=^\$ -cpu 1 -benchmem -bench '.*'                                                                         20:38 30/11
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_string_len                    	1000000000	         0.532 ns/op	       0 B/op	       0 allocs/op
Benchmark_string_len_reverse            	772742073	         1.56 ns/op	       0 B/op	       0 allocs/op
Benchmark_string_struct_len_reverse     	761126870	         1.55 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get/same_keys     	11334487	       113 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get/same_but_fresh_keys         	 7499438	       166 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get/same,_fresh_and_random_order_keys         	 4416318	       286 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys                                     	 9499071	       154 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_but_fresh_keys                           	 7277200	       174 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same,_fresh_and_random_order_keys             	 4206589	       313 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/static_key                      	92947183	        13.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_builtin_map_get_not_found/dynamic_keys                    	 9785622	       128 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1060 are full
Benchmark_simdmap_get_not_found/static_key                          	100000000	        11.8 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get_not_found/dynamic_keys                        	14061108	        94.3 ns/op	       0 B/op	       0 allocs/op
```

For ` ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*not_found/dynamic_keys.*'` Instruments shows:

```
L2 Hit:               ~8mm for simdmap (vs. 3mm for builtin)
DTLB walk completed:  ~10.2mm for simdmap (vs. 10.5 for builtin)
DTLB STLB hit:        ~3.3mm for simdmap (vs. 9.9mm for builtin)
Mispredicted branches ~0.9mm for simdmap (vs. 12mm for builtin)
```
seems like builtin just requests more data for not found.

For `-test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*get/same_but_fresh_keys.*` Intruments reports
```
L2 miss:               ~38.6mm for simdmap (vs. 35.4 for builtin)
Cycles:                ~4bn for simdmap (vs. 3.9bn for builtin)
Stalled cycles:        ~11.5bn for simdmap (vs. 9.9bn for builtin)
Mispredicted branches: ~0.4mm for simdmap (vs. 9.5mm for builtin)
```
Thus, simdmap is much more predictable for branch predictions, but the memory fetches seems to be worse.

Byte-aligned map
----------------
Making the `bucket` byte-aligned (by having a filler byte array) seems to improve performance of the simdmap

Instruments command
```
-test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 30000000x -test.cpuprofile cpu.profile -test.bench .*get/same_keys.*
```

Non-byte aligned:
```
testing: open cpu.profile: read-only file system
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	30000000	       114 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	30000000	       143 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1112 are full
PASS
```

Byte-aligned:
```
testing: open cpu.profile: read-only file system
goos: darwin
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	30000000	       110 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	30000000	       134 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1115 are full
PASS
<End of Run>
```

Comparing Instrument counters:
Non-byte aligned
```
Cpu cycles stalled:    33.867.820.180 for simdmap (vs. 24.266.816.843 for builtin)
L2 Cache misses:          112.878.397 for simdmap (vs. 98.156.923 for builtin)
```

Byte-aligned
```
Cpu cycles stalled:    29.347.750.948 for simdmap (vs. 25.248.233.827 for builtin)
L2 Cache misses:          95.792.491 for simdmap (vs. 98.296.924 for builtin)
```
So byte aligned does matter.


Next up?

With a byte-aligned map, the simdmap implementation for `get/same_keys` has fewer L2D misses, fewer TLB load misses causing a walk, but higher number of instructions that causes a stall:

```
Cycles with a stall: 25.474.354.712 (simdmap) vs. 20.477.061.167 (builtin)
L2 misses:           96.441.083 (simd) vs. 96.441.083 (builtin)
DTLB load misses:    31.015.416	(simd) vs. 34.998.766	(builtin)
```
So, I need to figure out what is causing these load stalls.


Turns out that DTLB is the first-level TBL cache, and STLB is the second-level. Looking at STLB, we see that simdmap does 40% more STLB misses than builtin. The `MEM_INST_RETIRED.STLB_MISS_LOADS` counts this

Args:
```
-test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 30000000x -test.bench .*get/same_keys.*
```
See Intel CPU events: https://download.01.org/perfmon/index/skylake.html
TLB numbers
```
Load instructions:       1.463.458.848 for SIMD (vs. 1.881.617.539 for builtin)
Loads with an STLB miss: 	  28.976.282 for SIMD (2% of loads - vs. 	21.631.208 for builtin - 1.1%)
DLTB Load misses:           31.020.833 for SIMD (vs. 35.153.593 for builtint)
DLTB miss, but in STLB:      2.789.990 for SIMD (vs. 30.226.279)
```
It thus seems like SIMD has more TLB misses. With `perf` you can see the instructions where those TLBs are, so I should somehow see if I can
get a Linux box to test perf on.

My current thesis is the more bound checks in the simd code, causes more TLB cache invalidations.



2021-01-08 Debugging TLB misses for get tests
=============================================

Not-found tests
---------------

For the not-found tests, TLB misses are the same for both versions:
```
$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*builtin.*not_found/dynamic_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get_not_found/dynamic_keys         	10000000	       168 ns/op	       0 B/op	       0 allocs/op
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*builtin.*not_found/dynamic_keys.*':

     2,278,803,379      dTLB-loads:u
        12,915,360      dTLB-load-misses:u        #    0.57% of all dTLB cache hits  # <- check this number and compare below, both are at 0.5% TLB misses
   <not supported>      dTLB-prefetch-misses:u

       6.397644517 seconds time elapsed

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*not_found/dynamic_keys.*'
Buckets: 131072, of which 1060 are full
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_simdmap_get_not_found/dynamic_keys         	10000000	       117 ns/op	       0 B/op	       0 allocs/op
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*simd.*not_found/dynamic_keys.*':

     2,474,095,181      dTLB-loads:u
        13,366,113      dTLB-load-misses:u        #    0.54% of all dTLB cache hits    # <- check this number and compare above, both are at 0.5% TLB misses
   <not supported>      dTLB-prefetch-misses:u

       5.389249197 seconds time elapsed
```

Found tests
-----------
And it seems that found tests are also better in the simdmap case (at least once):

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	10000000	       173 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	10000000	       131 ns/op	       0 B/op	       0 allocs/op # <- Simdmap is faster!!!!!
Buckets: 131072, of which 1083 are full
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*get/same_keys.*':

     8,018,199,273      dTLB-loads:u
        30,158,881      dTLB-load-misses:u        #    0.38% of all dTLB cache hits
   <not supported>      dTLB-prefetch-misses:u

      16.724325756 seconds time elapsed

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*builtin.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	10000000	       132 ns/op	       0 B/op	       0 allocs/op
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*builtin.*get/same_keys.*':

     3,547,374,496      dTLB-loads:u
        16,341,849      dTLB-load-misses:u        #    0.46% of all dTLB cache hits
   <not supported>      dTLB-prefetch-misses:u

       7.503344017 seconds time elapsed

$ ./linux-perf stat  -e dTLB-loads,dTLB-load-misses,dTLB-prefetch-misses ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_simdmap_get/same_keys 	10000000	       133 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1065 are full
PASS

 Performance counter stats for './simdmap.test -test.run=^$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench .*simd.*get/same_keys.*':

     3,870,569,488      dTLB-loads:u
        14,451,216      dTLB-load-misses:u        #    0.37% of all dTLB cache hits
   <not supported>      dTLB-prefetch-misses:u

       8.990690540 seconds time elapsed

Furthermore, line 179 in map.go (`if (keyP == ekeyP || keyP.Str == ekeyP.Str || z.Memequal(ekeyP.Str, keyP.Str, uintptr(keyP.Len))`) is causing a relative high number of TLB misses. It accounts for 31% of the TLB-loads, but accounts for 68% of the dTLB-load-misses. Similarly line 155 (`idx := __CompareNMask(grpCtrlPointer, unsafe.Pointer(hash>>57))`) accounts for 1.6% of the loads, but 16% of the dTLB-load-misses. I'll try to add
```
			if keyP.Len != ekeyP.Len {
				continue
			}
```
before line 179 to see if that changes the load-misses.

With the above, the above now has 30% of the dtlb-loads and 70%(!) of the dtlb-load-misses. Thus referencing `keyP.Len` or `ekeyP.Len` results in TLB misses. Based on my reading of the generated code the culprit is the `ekeyP := (*z.StringStruct)(unsafe.Pointer(&grp.keys[i]))` that generates a bunch of extra instructions, and the question is whether these instructions mess up the TLB cache. Should I try to write these as an `unsafe` array computation?

Using unsafe array computations
-------------------------------
Interesting, replacing `ekeyP := (*z.StringStruct)(unsafe.Pointer(&grp.keys[i]))` with `ekeyP := (*z.StringStruct)(add(unsafe.Pointer(&grp.keys), uintptr(i)*stringSize))`, keeps the tlb-misses at a LEA (load-effective address) instruction (see below). Thus, it seems that `&grp.keys` causes TLB-misses...
```
 map.go:175    0.99 :     5b0d54:       lea    (%rcx,%rdx,1),%rdi
    0.00 :        5b0d58:       lea    0x10(%rdi),%rdi
    0.05 :        5b0d5c:       mov    %rsi,%rax
    0.24 :        5b0d5f:       shl    $0x4,%rsi
    0.24 :        5b0d63:       mov    0x78(%rsp),%r8
    0.00 :        5b0d68:       mov    0x8(%rdi,%rsi,1),%r9
         :      github.com/jakobgt/cay.add():
 map.go:242   68.16 :     5b0d6d:       lea    (%rdi,%rsi,1),%r10 # <- this one in the add function
         :      github.com/jakobgt/cay.(*Map).find():
    0.00 :        5b0d71:       cmp    %r8,%r9
    0.00 :        5b0d74:       jne    5b0d39 <github.com/jakobgt/cay.(*Map).find+0x79>
```

Hence next step is to figure out whether I really need `grp := &m.buckets[cGroup]` or it should be `grp := m.buckets[cGroup]`


2020-01-11 Stack-allocating the bucket
======================================

If using `grp := m.buckets[cGroup]`, I would have hoped that the bucket would have been allocated on the stack. From the generated byte code, it does not seem to be case, as `runtime.newobject` is called along with `runtime.duffcopy` and ` runtime.typedmemmove`, so maybe the Go compiler thinks that the bucket escapes and thus is allocated on the heap. (`newobject` is the same as malloc).

```
 155            .      2.17s           		grp := m.buckets[cGroup]
                   .          .  11e58b0:                     LEAQ runtime.rodata+297792(SB), AX                           map.go:155
                   .          .  11e58b7:                     MOVQ AX, 0(SP)                                               map.go:155
                   .          .  11e58bb:                     NOPL 0(AX)(AX*1)                                             map.go:155
                   .      1.13s  11e58c0:                     CALL runtime.newobject(SB)                                   map.go:155
 ...
 (AX*1)                                             map.go:155
                   .          .  11e58e0:                     CMPQ CX, BX                                                  map.go:155
                   .          .  11e58e3:                     JAE 0x11e5995                                                map.go:155
                   .          .  11e58e9:                     MOVQ DI, 0x48(SP)                                            map.go:155
                   .          .  11e58ee:                     SHLQ $0xa, BX                                                map.go:155
                   .          .  11e58f2:                     LEAQ 0(DX)(BX*1), SI                                         map.go:155
                   .          .  11e58f6:                     CMPL $0x0, runtime.writeBarrier(SB)                          map.go:155
                   .          .  11e58fd:                     JNE 0x11e596f                                                map.go:155
                   .          .  11e58ff:                     NOPL                                                         map.go:155
                   .          .  11e5900:                     MOVQ BP, -0x10(SP)                                           map.go:155
                   .          .  11e5905:                     LEAQ -0x10(SP), BP                                           map.go:155
                   .      280ms  11e590a:                     CALL runtime.duffcopy(SB)                                    map.go:155
                   .          .  11e590f:                     MOVQ 0(BP), BP                                               map.go:155
...
                   .      760ms  11e5984:                     CALL runtime.typedmemmove(SB)                                map.go:155
```

Btw., `go build -gcflags '-m' . ` is good for seeing what escapes to the heap. By ensuring that the `grp` variable does not escape, I reduce the ns/ops from 1100ns to 200ns. And thus, no mallocs and `typedmemmove`. Only a `duffcopy` call.

Making sure that `grp` does not escape, we see a different picture, where the tlb-misses are located in (percentages)
```
 47.10 map.go:138
   14.73 map.go:231
   14.66 map.go:155
   10.13 map.go:153
    5.17 map.go:236
    3.97 map.go:151
    2.34 map.go:150
```
where
 - `map.go:138` is `hash := z.Memhash(keyP.Str, 0, uintptr(keyP.Len))`
 - `map.go:231` is `uintptr(1) << (b & (PtrSize*8 - 1))`, where `b` is `m.logSize`, so I guess a lookup of the `m` variable
 - `map.go:155` is `grp := m.buckets[cGroup]`

So maybe the duffcopy is destroying the TLB cache? Looking in the duffcopy code, it has 14610 samples, whereas the `find` method has 1412 samples. Thus duffcopy seems to kill the TLB. Thus, can we load the TLB for the full page of `grp`, without a `duffcopy` call? It would be interesting to see whether the `grp` is page aligned? Or rather can we just copy the control byte array, which causes a copy of the array, but a full load of the page into the TLB?

By forcing it to copy the control byte array onto the stack, I've seen some movups, thus the compiler can't see that the data is byte-aligned (MOVUPS is Move Unaligned Packed Single-Precision). Furthermore copying a byte array is not the same as copying a string array (the latter leads to a duffcopy call).

Confirmed with the test case `Test__FirstBucketIsPageAligned`, that the buckets are page-aligned (as in the first bucket is). Thus, the TLB misses are not coming from misaligned loads. Thus need to figure out why there is a TLB miss, when accessing `unsafe.Pointer(&grp.keys[i])`.

Maybe the second TLB miss, which stands for 13% of the misses in `find()` is okay? Can I find optimizations somewhere else?

2020-01-13 Caching the returned pointer
=======================================
It seems that we can't optimize `find` that much more, where we have to accept the second round of TLB misses, when inspecting the length of the keys. Maybe we can optimize the case where the key is found and then making sure that, once we return the value pointer we don't hit a third TLB miss there?

Inspecting the code with
```
./perf record -e cache-misses -g ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*get/same_keys.*'
```
it seems that all the cache misses happens where there are also TLB misses, including the return call in Get().

Two ideas: a) Move key and value closer, and/or b) reduce function calls and LOC between returning from find and using the byte[]

a) Idea: pair key and value together.
-------------------------------------

With
```
type entry struct {
	key   string
	value []byte
}
```
for each entry it seems that we get a lot fewer TLB misses in the return statement (~248 samples, whereas the ekeyP.Len generates ~1K TLB page misses). This could be interesting.

As a side note, SIMD get/same keys is often times, faster than the builtin:
```
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 30s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	199967498	       194 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	202134694	       184 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1043 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	86886939	       143 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	100000000	       125 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1081 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	84544454	       149 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	100000000	       124 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1120 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	94161273	       154 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	100000000	       172 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1109 are full
PASS
$ ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10s -test.bench '.*get/same_keys.*'
goos: linux
goarch: amd64
pkg: github.com/jakobgt/cay
Benchmark_builtin_map_get/same_keys 	93228235	       145 ns/op	       0 B/op	       0 allocs/op
Benchmark_simdmap_get/same_keys     	98207223	       122 ns/op	       0 B/op	       0 allocs/op
Buckets: 131072, of which 1106 are full
PASS
```

2020-01-14: Look into the cache misses
======================================

Generally, where the code is using the most cpu-cycles is also where there are cache misses, so I want to know what type of cache miss it is. E.g., TLB/L1/L2, etc. On the host, the L1 is 32K.

Running
```
./perf record -e L1-dcache-load-misses -g ./simdmap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*get/same_keys.*'
```
gives that `if keyP.Len != ekeyP.Len {` (`map_get:75`) has:
- 80% of the L1 cache misses in `find_get` (2381 samples)
- 70% of the LLC cache misses. (LLC = Last Level Cache, not sure whether that is L2 or L3LLC-load-misses) (2353 samples)

`__CompareNMask` has a similar level of L1 cache misses, but not LLC misses (only 600 samples).

This is really weird, as the host has a 32KB L1 cache, 256KB L2 cache and a 20MB L3 cache and that specific line gets flushed. Could it be the GC that flushes the caches?



TODO: Try to disable GC during the test runs. That might avoid the GC. Look for deltablue benchmarks, that were used in V8/Dart, etc.

2021-03-30: Try memory ballast
==============================
Adding some memory ballast in the form of
```
var (
	_ballast = make([]byte, int64(1)<<int64(33)) // 8GB
)
```

does not change anything unfortunately.

Changing the number of entries in the map makes a big difference, as for all maps smaller than `1<<15`, simdmap is faster on both Mac and Linux. And it is faster always.

Maybe the type should be changes to `map<string, int>`, which is what Matt Kulundis from Google is using to benchmark his types. He has two set-ups, one with a 4 byte key/value and one with 64 bytes (he is only looking at sets, where the key and value is the same).
