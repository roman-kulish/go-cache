package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const (
	maxKeyLen  = 32
	maxDataLen = 1024

	Records     = 10000
	Concurrency = 10
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	var u string
	var n, c uint

	flag.StringVar(&u, "url", "", "Cache server URL")
	flag.UintVar(&n, "num", Records, "Number of records to create")
	flag.UintVar(&c, "con", Concurrency, "Concurrency, a number of goroutine")
	flag.Parse()

	if u == "" {
		log.Fatal("URL is not specified")
	}

	if n == 0 {
		log.Printf("WARNING: number of records must be greater than zero, using default %d", Records)

		n = Records
	}

	if c == 0 {
		log.Printf("WARNING: number of goroutine must be greater than zero, using default %d", Concurrency)

		n = Records
	}

	addr, err := url.Parse(u)

	if err != nil {
		log.Fatal(err)
	}

	seed(addr, n, c)
}

func seed(addr *url.URL, n, c uint) {
	var i uint

	sem := make(chan struct{}, c)
	chErr := make(chan error)
	wg:= sync.WaitGroup{}

	for ; i < n; i++ {
		select {
		case err := <-chErr:
			log.SetOutput(os.Stderr)
			log.Fatal(err) // exit if any worker returned error

		default:
		}

		addr.Path = string(randBytes(maxKeyLen))
		dest := addr.String()

		sem <- struct{}{} // book token
		wg.Add(1)

		go push(dest, randBytes(maxDataLen), sem, chErr, &wg)
	}

	close(sem)
	close(chErr)

	wg.Wait()
}

func push(dest string, p []byte, sem chan struct{}, chErr chan error, wg *sync.WaitGroup) {
	defer func() {
		<-sem // release token
		wg.Done()
	}()

	res, err := http.Post(dest, "text/plain", bytes.NewBuffer(p))

	if err == nil && res.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("server returned status code %d", res.StatusCode)
	}

	if err != nil {
		chErr <- err
		return
	}

	fmt.Println(http.MethodGet, dest)
}

func randBytes(len int) []byte {
	n := 1 + r.Intn(len)
	b := make([]byte, n)

	for i := 0; i < n; i++ {
		b[i] = byte(65 + r.Intn(25))
	}

	return b
}
