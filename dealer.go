package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type dealer interface {
	Run()
	NextBlock() []byte
	Found(string)
}

type dealerServer struct {
	pos []int8
	lock sync.Mutex
	standalone bool
}

func newDealerServer(standalone bool) *dealerServer {
	return &dealerServer{
		pos: make([]int8, tripLength),
		standalone: standalone,
	}
}

func (d *dealerServer) Run() {
	if d.standalone {
		return
	}

	http.HandleFunc("/pos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		pos := d.incrAndCopy()
		en := json.NewEncoder(w)
		err := en.Encode(pos)
		if err != nil {
			// any error should not happen here
			panic(err)
		}
	})

	http.HandleFunc("/found", func(w http.ResponseWriter, r *http.Request) {
		trip, err := url.PathUnescape(r.URL.Query().Get("trip"))
		if err != nil {
			fmt.Printf("\nDealer: failed to decode the trip of found request: %s\n", err)
			return
		}
		found(trip, r.URL.Query().Get("by"))
	})

	go func() {
		err := http.ListenAndServe("0.0.0.0:52313", nil)
		if err != nil {
			// any error should not happen here
			panic(err)
		}
	}()
	fmt.Println("Dealer is serving at 0.0.0.0:52313")
}

func (d *dealerServer) NextBlock() []byte {
	posc := d.incrAndCopy()
	buf := make([]byte, tripLength)
	for i := 0; i < tripLength; i++ {
		buf[i] = chars[posc[i]]
	}
	return buf
}

func (d *dealerServer) incrAndCopy() []int8 {
	d.lock.Lock()
	defer d.lock.Unlock()
	copied := append([]int8{}, d.pos...)

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
	return copied
}

func (d *dealerServer) Found(trip string) {
	found(trip, "myself")
}

type dealerClient struct {
	host string
	cli http.Client
	pos chan [tripLength]uint8
}

func newDealerClient(remoteHost string) *dealerClient {
	return &dealerClient{
		host: remoteHost,
		cli: http.Client{Timeout: time.Second},
		pos: make(chan [tripLength]uint8, 1),
	}
}

func (d *dealerClient) Run() {
	go func() {
		for {
			err := d.get()
			if err != nil {
				fmt.Printf("\nClient: %s\n", err)
			}
		}
	}()
}

func (d *dealerClient) get() error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:52313/pos", d.host), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	res, err := d.cli.Do(req)
	if err != nil {
		return fmt.Errorf("failed to GET: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200: %d", res.StatusCode)
	}

	pos := [tripLength]uint8{}
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&pos)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %s", err)
	}

	d.pos <- pos
	return nil
}

func (d *dealerClient) NextBlock() []byte {
	pos := <- d.pos
	buf := make([]byte, tripLength)
	for i := 0; i < tripLength; i++ {
		buf[i] = chars[pos[i]]
	}
	return buf
}

func (d *dealerClient) Found(trip string) {
	found(trip, "myself")

	host, err := os.Hostname()
	if err != nil {
		host = "unnamed-client"
	}

	u := fmt.Sprintf("http://%s:52313/found?by=%s&trip=%s", d.host, host, url.PathEscape(trip))
	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		fmt.Printf("\nClient: failed to create found request: %s\n", err)
		return
	}

	res, err := d.cli.Do(req)
	if err != nil {
		fmt.Printf("\nClient: failed to post found request: %s\n", err)
		return
	}

	if res.StatusCode != http.StatusOK {
		fmt.Printf("\nClient: remote server returned non-200 response for the found request: %d\n", res.StatusCode)
	}
}

func found(trip, by string) {
	sha := sha1.New()
	sha.Write([]byte(trip))
	s := base64.StdEncoding.EncodeToString(sha.Sum(nil))
	fmt.Printf("\nFOUND!!!: #%s -> %s (by %s)\n", trip, s, by)
}
