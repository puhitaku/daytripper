package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	tripLength = 16
	charsLen = 65
)

var chars []byte = []byte("0123456789abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+|[];',./{}:\"<>?`~")

func main() {
	n := flag.Int("nr", runtime.NumCPU()*2, "Number of goroutines (default: runtime.NumCPU() * 2)")
	remote := flag.String("remote", "", "Remote daytripper host (optional for distributed calculation)")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Invalid number of args!")
		return
	}

	prefix := flag.Arg(0)
	fmt.Printf("Searching for '%s' with %d goroutines...\n", prefix, *n)

	var d dealer

	if *remote == "" {
		d = newDealerServer()
	} else {
		d = newDealerClient(*remote)
	}

	d.Run()

	ts := make([]*tripper, *n)
	eg := errgroup.Group{}

	for i := 0; i < *n; i++ {
		j := i
		ts[j] = newTripper(d)
		eg.Go(func() error {
			return ts[j].Go(prefix, false)
		})
	}

	start := time.Now()

	go func() {
		var count uint64
		var lastCount uint64
		for {
			lastCount = count
			count = 0
			for i := 0; i < *n; i++ {
				count += ts[i].Count
			}
			fmt.Printf("Hashes: %d (%d hash/s) | Elapsed %d sec", count, count-lastCount, time.Now().Sub(start) / time.Second)
			time.Sleep(time.Second)
			fmt.Print("\r")
		}
	}()

	err := eg.Wait()
	if err != nil {
		panic(err)
	}
}

