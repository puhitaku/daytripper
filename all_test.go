package main

import "testing"

var d *dealerServer
var t *tripper

func Benchmark(b *testing.B) {
	d = newDealerServer(true)
	t = newTripper(d, tripperConfig{
		Prefix: "aaaaaaaaaa",
		Once:   true,
	})

	b.Run("BenchmarkTripper_GoOne", benchmarkGoOne)
}

func benchmarkGoOne(b *testing.B) {
	err := t.Go()
	if err != nil {
		panic(err)
	}
}
