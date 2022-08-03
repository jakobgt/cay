# cay
SIMD-based hashmap for Go.

To start:
---------

- go get -u github.com/minio/asm2plan9s

Developing
----------

To rebuild the ASM files from the c code:

    ./build.sh

Running
-------
To run the benchmarks:

    go test -run=^\$ -cpu 8 -bench '.*' --benchmem -cpuprofile cpu.profile

Checking the profile

    go tool pprof   --http :1234 simdmap.test cpu.profile

Debugging
---------
 - `go tool objdump -s find simdmap.test` inspecting bytecode

 - `go build -gcflags '-m -m' .`


TODO
----

- Get the tests to be green again.
- Compare memory footprint of simdmap and builtin.
- `HighestBitMask` has a tendency to crash, unless `ctrlGroup` is allocated on the heap. My best guess is that it is otherwise not stack aligned (some architectures require 16-byte aligned stacks).
- Returning the element in `Get` is costing 33% of the overall runtime and might be some locality or copying that is expensive.


Observations
------------

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



References
== HashBrown and SwissTable
 - Run-through of the SIMD hash table: https://blog.waffles.space/2018/12/07/deep-dive-into-hashbrown/
 - Explanation of the rust version of SwissTable: https://gankra.github.io/blah/hashbrown-insert/
 - Code of the Simd Map implementation in Rust: https://sourcegraph.com/github.com/Amanieu/hashbrown@master/-/blob/src/raw/mod.rs#L444:19

== Bit operations
 - Fast bitset iteration: https://lemire.me/blog/2018/02/21/iterating-over-set-bits-quickly/

== Calling ASM in Go
 - Optimizing stuff: https://medium.com/@c_bata_/optimizing-go-by-avx2-using-auto-vectorization-in-llvm-118f7b366969
 - Intrinsics in Golang: https://dave.cheney.net/2019/08/20/go-compiler-intrinsics
