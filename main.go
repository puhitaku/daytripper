package main

import (
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"hash"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	tripLength = 20
	charsLen = 65
	blockSize uint64 = charsLen*charsLen*charsLen*charsLen
)

var chars []byte = []byte("0123456789abcdefghijklmnopqrstuvwxyz!@#$%^&*()_+|[];',./{}:\"<>?`~")

type dealer struct {
	l int
	pos []uint8
	buf [][tripLength]byte
	lock sync.Mutex

	Count uint64
}

func newDealer(n int) *dealer {
	return &dealer {
		l: tripLength,
		pos: make([]uint8, tripLength),
		buf: make([][tripLength]byte, n),
	}
}

func (d *dealer) NextBlock(rn int) []byte {
	posc := d.incrAndCopy()
	for i := 0; i < tripLength; i++ {
		d.buf[rn][i] = chars[posc[i]]
	}
	return d.buf[rn][:]
}

func (d *dealer) incrAndCopy() []uint8 {
	d.lock.Lock()
	defer d.lock.Unlock()
	copied := append([]byte{}, d.pos...)

	for i := 4; i < tripLength+1; i++ {
		if i == tripLength {
			panic("limit exceeded!")
		}

		if d.pos[i] < charsLen-1 {
			d.pos[i] += 1
			break
		}

		d.pos[i] = 0
	}

	d.Count++
	return copied
}

type tripper struct {
	i int
	h hash.Hash
	d *dealer

	Count uint64
}

func newTripper(d *dealer, i int) *tripper {
	return &tripper{
		i: i,
		h: sha1.New(),
		d: d,
	}
}

func (t *tripper) Go(prefix string) error {
	var bufi []byte
	var bufo []byte = make([]byte, charsLen*2)

	for i := 0; ; i++ {
		bufi = t.d.NextBlock(t.i)
		for j := 0; j < int(blockSize); j++ {
			bufi[0] = chars[j % charsLen]
			bufi[1] = chars[(j / charsLen) % charsLen]
			bufi[2] = chars[(j / charsLen / charsLen) % charsLen]
			bufi[3] = chars[(j / charsLen / charsLen / charsLen) % charsLen]
			t.h.Reset()
			t.h.Write(bufi)
			base64.StdEncoding.Encode(bufo, t.h.Sum(nil))
			if strings.HasPrefix(string(bufo), prefix) {
				fmt.Printf("\rFOUND!!!: #%s -> %s\n", string(bufi), strings.TrimRight(string(bufo), "\x00"))
			}

			t.Count++
		}
	}
}

func (t *tripper) GoOne(prefix string) error {
	var bufo = make([]byte, charsLen*2)

	bufi := t.d.NextBlock(t.i)
	for j := 0; j < int(blockSize); j++ {
		bufi[0] = chars[j % charsLen]
		bufi[1] = chars[(j / charsLen) % charsLen]
		bufi[2] = chars[(j / charsLen / charsLen) % charsLen]
		bufi[3] = chars[(j / charsLen / charsLen / charsLen) % charsLen]

		t.h.Reset()
		t.h.Write(bufi)
		base64.StdEncoding.Encode(bufo, t.h.Sum(nil))
		if strings.HasPrefix(string(bufo), prefix) {
			fmt.Printf("\rFOUND!!!: #%s -> %s\n", string(bufi), strings.TrimRight(string(bufo), "\x00"))
		}

		t.Count++
	}
	return nil
}

func main() {
	n := flag.Int("nr", runtime.NumCPU(), "Number of goroutines (default: runtime.NumCPU())")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("Invalid number of args!")
		return
	}
	prefix := flag.Arg(0)

	fmt.Printf("Searching for '%s' with %d goroutines...\n", prefix, *n)

	d := newDealer(*n)
	ts := make([]*tripper, *n)
	eg := errgroup.Group{}

	for i := 0; i < *n; i++ {
		j := i
		ts[j] = newTripper(d, j)
		eg.Go(func() error {
			return ts[j].Go(prefix)
		})
	}

	go func() {
		var count uint64
		var lastCount uint64
		for {
			lastCount = count
			count = 0
			for i := 0; i < *n; i++ {
				count += ts[i].Count
			}
			fmt.Printf("Hashes: %d (%d hash/s)", count, count-lastCount)
			time.Sleep(time.Second)
			fmt.Print("\r")
		}
	}()

	err := eg.Wait()
	if err != nil {
		panic(err)
	}
}

