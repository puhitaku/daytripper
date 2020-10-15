package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash"
	"math"
	"strings"
)

type tripperConfig struct {
	Prefix string
	Once bool
}

type tripper struct {
	h hash.Hash
	d dealer
	conf tripperConfig

	Count uint64
}

func newTripper(d dealer, conf tripperConfig) *tripper {
	return &tripper{
		h: sha1.New(),
		d: d,
		conf: conf,
	}
}

func (t *tripper) Go() error {
	prefix := t.conf.Prefix

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
	if t.conf.Once {
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
