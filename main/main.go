package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jakobgt/cay"
	"github.com/pkg/profile"
)

func main() {
	size := 1 << 10
	iterations := 1
	if len(os.Args) > 1 {
		s, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("%s is not a valid size. should be an int", os.Args[1])
			return
		}
		size = s
	}

	if len(os.Args) > 2 {
		iter, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("%s is not count number. should be an int", os.Args[2])
			return
		}
		iterations = iter
	}

	// We initialize
	fmt.Println("Warming up")
	keys := cay.RandomKeys(size)
	m := cay.Simdmap(keys)
	readAll(m, keys)
	fmt.Println("Running test")
	pr := profile.Start(profile.CPUProfile, profile.ProfilePath("."))
	start := time.Now()
	for i := 0; i < iterations; i++ {
		readAll(m, keys)
	}
	timing := time.Now().Sub(start)
	pr.Stop()

	ops_per_sec := timing.Nanoseconds() / (int64(iterations) * int64(size))
	fmt.Printf("Total time: %dns\n", timing.Nanoseconds())
	fmt.Printf("Size: %d, Iterations: %d\n", size, iterations)
	fmt.Printf("Ops per seco %d ns\n", ops_per_sec)
	//for _, k := range keys {
	//	m.Get(k)
	//}
	//
	//// Now we measure
}

func readAll(m *cay.Map, keys []string) {
	for _, k := range keys {
		v, ok := m.Get(k)
		if !ok {
			fmt.Println(v)
		}
	}
}
