# cay - SIMD-based hashmap for Go.

> Note that this is an experimental prototype, so don't use in production

Cay's goal is to be a faster and more memory-efficient replacements for the builtin map in Go.

Cay is heavily inspired by the [Hashbrown](https://blog.waffles.space/2018/12/07/deep-dive-into-hashbrown/)
map in Rust (which in turn is a port of the [Swisstable](https://www.youtube.com/watch?v=ncHmEUmJZf4))

As this is my first real endevaour in the world of memory operations and optimizations in Go, I have included my
log of my (mostly failed) experimentation.


## Developing on caymap


### To start:

- go get -u github.com/minio/asm2plan9s

### Developing

To rebuild the ASM files from the c code:

    ./build.sh

### Running
To run the benchmarks:

    go test -run=^\$ -cpu 1 -bench '.*' --benchmem

Checking the profile

    go tool pprof   --http :1234 caymap.test cpu.profile

### Debugging
 - `go tool objdump -s find caymap.test` inspecting bytecode

 - `go build -gcflags '-m -m' .`

### Using perf

Build the caymap tests for linux:
- `env GOOS=linux GOARCH=amd64 go test -c -o caymap.test`

Record the instructions:
```
./perf record -e dTLB-load-misses -g ./caymap.test -test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*simd.*get/same_keys.*'
```

View

```
./perf annotate --stdio -l --tui
```

### Using Instruments
To use Instruments, first compile the test binary:

```
# From the simdmap dir
go test -c -o caymap.test
```

Then open instruments and use the `caymap.test` file as the target. For arguments you can pass in
```
-test.run=^\$ -test.cpu 1 -test.benchmem -test.benchtime 10000000x -test.bench '.*not_found/dynamic_keys.*'
```

and that will run each the not found benchmarks 10,000,000 times. Change the regexp for the other benchmarks.

## TODO

- Compare memory footprint of caymap and builtin.
- `HighestBitMask` has a tendency to crash, unless `ctrlGroup` is allocated on the heap. My best guess is that it is otherwise not stack aligned (some architectures require 16-byte aligned stacks).
- Returning the element in `Get` is costing 33% of the overall runtime and might be some locality or copying that is expensive.
- Investigate tuning huge pages and transparent huge pages on.
- Investigate page alignment

## References

### HashBrown and SwissTable
 - [Run-through of the SIMD hash table](https://blog.waffles.space/2018/12/07/deep-dive-into-hashbrown/)
 - [Explanation of the rust version of SwissTable](https://gankra.github.io/blah/hashbrown-insert/)
 - [Code of the Simd Map implementation in Rust](https://sourcegraph.com/github.com/Amanieu/hashbrown@master/-/blob/src/raw/mod.rs#L444:19)
 - [Google Presentation on SwissTable](https://www.youtube.com/watch?v=ncHmEUmJZf4) - code is in https://github.com/abseil/abseil-cpp/blob/master/absl/container/internal/raw_hash_set.h

### Bit operations
 - [Fast bitset iteration](https://lemire.me/blog/2018/02/21/iterating-over-set-bits-quickly/)

### Calling ASM in Go
 - [Optimizing stuff](https://medium.com/@c_bata_/optimizing-go-by-avx2-using-auto-vectorization-in-llvm-118f7b366969)
 - [Intrinsics in Golang](https://dave.cheney.net/2019/08/20/go-compiler-intrinsics)
