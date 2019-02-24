package cache

import (
	"fmt"
	"hash"
	"hash/fnv"
)

type NewCacheFunc func(capacity uint) Cache

type shardedCache struct {
	shards  []Cache
	nShards uint32
	hash    hash.Hash32
}

func (s *shardedCache) shard(key string) (hashed uint32, err error) {
	s.hash.Reset()

	if _, err = s.hash.Write([]byte(key)); err != nil {
		return
	}

	return s.hash.Sum32() % s.nShards, nil
}

func (s *shardedCache) Set(key string, value []byte) error {
	shard, err := s.shard(key)

	if err != nil {
		return err
	}

	return s.shards[shard].Set(key, value)
}

func (s *shardedCache) Get(key string) ([]byte, error) {
	shard, err := s.shard(key)

	if err != nil {
		return nil, fmt.Errorf("error computing shard ID for key [%s]: %s", key, err)
	}

	return s.shards[shard].Get(key)
}

func NewSharded(shards uint8, capacity uint, f NewCacheFunc) Cache {
	var i uint8

	s := make([]Cache, shards)

	for ; i < shards; i++ {
		s[i] = f(capacity)
	}

	return &shardedCache{
		shards:  s,
		nShards: uint32(shards),
		hash:    fnv.New32a(),
	}
}
