package main

import (
	"flag"
	"fmt"
	"os"
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
	flag.Usage = usage

	nr := flag.Int("nr", runtime.NumCPU() * 2, "Number of goroutines (default: runtime.NumCPU() * 2)")
	remote := flag.String("remote", "", "Remote daytripper host (optional for distributed calculation)")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Invalid number of args!")
		return
	}

	prefix := flag.Arg(0)
	fmt.Printf("Searching for '%s' with %d goroutines...\n", prefix, *nr)

	var d dealer

	if *remote == "" {
		d = newDealerServer()
	} else {
		d = newDealerClient(*remote)
	}

	d.Run()

	ts := make([]*tripper, *nr)
	eg := errgroup.Group{}

	for i := 0; i < *nr; i++ {
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
			for i := 0; i < *nr; i++ {
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

var usageStr = `
Usage: %s [-nr N] [-remote HOST] TRIP
  TRIP
        The trip substring to find
  -nr N (int)
        Number of goroutines (default: runtime.NumCPU() * 2)
  -remote HOST (string)
        Remote daytripper host (optional for distributed calculation)
`

func usage() {
	fmt.Printf(usageStr[1:], os.Args[0])
}

