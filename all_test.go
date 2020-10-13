package main

import "testing"

var d *dealerServer
var t *tripper

func Benchmark(b *testing.B) {
	d = newDealerServer()
	t = newTripper(d)

	b.Run("BenchmarkTripper_GoOne", benchmarkGoOne)
}

func benchmarkGoOne(b *testing.B) {
	err := t.Go("aaaaaaaaaa", true)
	if err != nil {
		panic(err)
	}
}
