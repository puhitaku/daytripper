package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"hash"
	"math"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	tripLength = 16
	charsLen = 65
	blockSize uint64 = charsLen*charsLen*charsLen*charsLen
)

var chars []byte = []byte("0123456789abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+|[];',./{}:\"<>?`~")

type tripper struct {
	h hash.Hash
	d Dealer

	Count uint64
}

func newTripper(d Dealer) *tripper {
	return &tripper{
		h: sha1.New(),
		d: d,
	}
}

func (t *tripper) Go(prefix string, once bool) error {
	if len(prefix) < 5 {
		return fmt.Errorf("too short")
	}

	prefixp := prefix
	if len(prefix)%4 != 0 {
		prefixp += strings.Repeat(prefix[len(prefix)-1:], 4-len(prefix)%4)
	}
	expect, err := base64.StdEncoding.DecodeString(prefixp)
	if err != nil {
		return fmt.Errorf("failed to decode prefix: %s", err)
	}
	expect = expect[:len(expect)-3]  // the last 18 bits (3 bytes) can have different byte than we expect

	var bufi []byte
	var bufo []byte = make([]byte, charsLen*2)
	prefixb := []byte(prefix)

	iLimit := uint64(math.MaxUint64)
	if once {
		iLimit = 1
	}

	for i := uint64(0); i < iLimit; i++ {
		bufi = t.d.NextBlock()
		for j1 := 0; j1 < charsLen; j1++ {
			bufi[0] = chars[j1]
			for j2 := 0; j2 < charsLen; j2++ {
				bufi[1] = chars[j2]
				for j3 := 0; j3 < charsLen; j3++ {
					bufi[2] = chars[j3]
					for j4 := 0; j4 < charsLen; j4++ {
						bufi[3] = chars[j4]

						t.h.Reset()
						t.h.Write(bufi)
						if bytes.HasPrefix(t.h.Sum(nil), expect) {
							base64.StdEncoding.Encode(bufo, t.h.Sum(nil))
							if bytes.HasPrefix(bufo, prefixb) {
								t.d.Found(string(bufi))
							}
						}

						t.Count++
					}
				}
			}
		}
	}

	return nil
}

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

	var d Dealer

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

