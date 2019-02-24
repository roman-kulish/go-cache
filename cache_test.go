package cache

import (
	"math/rand"
	"testing"
	"time"
)

const (
	maxStrLen = 32
	cacheCap  = 100
	shardsLen = 16
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func BenchmarkMapCache(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(bench(NewMap(cacheCap)))
}

func BenchmarkChannelCache(b *testing.B) {
	c := NewChannel(cacheCap)

	b.ReportAllocs()
	b.RunParallel(bench(c))

	c.(*channelCache).Close()
}

func BenchmarkBufferCache(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(bench(NewBuffer(cacheCap*(maxStrLen*3))))
}

func BenchmarkShardedMapCache(b *testing.B) {
	f := NewCacheFunc(func(capacity uint) Cache {
		return NewMap(capacity)
	})

	b.ReportAllocs()
	b.RunParallel(bench(NewSharded(shardsLen, cacheCap, f)))
}

func BenchmarkShardedBufferCache(b *testing.B) {
	f := NewCacheFunc(func(capacity uint) Cache {
		return NewBuffer(uint32(capacity))
	})

	b.ReportAllocs()
	b.RunParallel(bench(NewSharded(shardsLen, cacheCap*(maxStrLen*3), f)))
}

func bench(c Cache) func(*testing.PB) {
	var keys []string

	for i := 0; i < cacheCap; i++ {
		key := string(randBytes())
		keys = append(keys, key)

		_ = c.Set(key, randBytes())
	}

	return func(pb *testing.PB) {
		var i uint

		for pb.Next() {
			c.Get(keys[i])
		}

		if i+1 == cacheCap {
			i = 0
		} else {
			i++
		}
	}
}

func randBytes() []byte {
	n := 1 + r.Intn(maxStrLen)
	b := make([]byte, n)

	for i := 0; i < n; i++ {
		b[i] = byte(65 + r.Intn(25))
	}

	return b
}
