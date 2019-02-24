package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/roman-kulish/go-cache"
)

const (
	Map           = "map"
	Buffer        = "buffer"
	Channel       = "channel"
	ShardedMap    = "sharded_map"
	ShardedBuffer = "sharded_buffer"

	Capacity = 10000
	Shards   = 16
)

type handler struct {
	cache cache.Cache
}

func main() {
	var ct string
	var cp, sh uint
	var c cache.Cache

	flag.StringVar(&ct, "cache", Map, fmt.Sprintf("Cache type (%s, %s, %s, %s, %s)",
		Map,
		Buffer,
		Channel,
		ShardedMap,
		ShardedBuffer))

	flag.UintVar(&cp, "cap", Capacity, "Cache capacity")
	flag.UintVar(&sh, "shards", Shards, "Cache shards number (1-255)")
	flag.Parse()

	if cp == 0 {
		log.Printf("WARNING: cache capacity must be greater than zero, using default %d", Capacity)

		cp = Capacity
	}

	if sh == 0 || sh > 255 {
		log.Printf("WARNING: invalid shards number %d, using default %d", sh, Shards)

		sh = Shards
	}

	switch ct {
	case Map:
		c = cache.NewMap(cp)

	case Buffer:
		c = cache.NewBuffer(uint32(cp))

	case Channel:
		c = cache.NewChannel(cp)

	case ShardedMap:
		c = cache.NewSharded(uint8(sh), cp, func(capacity uint) cache.Cache {
			return cache.NewMap(capacity)
		})

	case ShardedBuffer:
		c = cache.NewSharded(uint8(sh), cp, func(capacity uint) cache.Cache {
			return cache.NewBuffer(uint32(capacity))
		})

	default:
		log.Fatalf("Unsupported cache type: %s", ct)
	}

	handler := handler{
		cache: c,
	}

	server := http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p []byte
	var err error

	key := strings.Trim(r.URL.Path, "/ ")
	key = strings.ToLower(key)

	if key == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		p, err = h.cache.Get(key)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			io.WriteString(w, err.Error())
			return
		}

		if p == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Write(p)

	case http.MethodPost:
		if p, err = ioutil.ReadAll(r.Body); err == nil {
			err = h.cache.Set(key, p)
		}

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			io.WriteString(w, err.Error())
			return
		}

		w.WriteHeader(http.StatusAccepted)

	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
