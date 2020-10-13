package main

import "testing"

var d *dealer
var t *tripper

func Benchmark(b *testing.B) {
	d = newDealer(1)
	t = newTripper(d, 0)

	b.Run("BenchmarkTripper_GoOne", benchmarkGoOne)
}

func benchmarkGoOne(b *testing.B) {
	err := t.Go("aaaaaaaaaa", true)
	if err != nil {
		panic(err)
	}
}
