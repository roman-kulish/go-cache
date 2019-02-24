package cache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"sync"
)

const (
	keyHdrSize = 8
	lenHdrSize = 4
)

var (
	ErrTooBig        = errors.New("data is too big")
	ErrKeyCollision  = errors.New("cache key collision")
	ErrOutOfBounds   = errors.New("cached data length is out-of-bounds")
	ErrCorruptedData = errors.New("cached data is corrupted")

	marker     = []byte("%REC%")
	markerSize = uint32(len(marker))
)

type bufferCache struct {
	index    map[uint64]uint32
	cache    []byte
	ptr      uint32
	capacity uint32
	mu       *sync.RWMutex
	hash     hash.Hash64
}

func (c *bufferCache) Set(key string, value []byte) (err error) {
	dataLen := uint32(len(value))
	recSize := keyHdrSize + lenHdrSize + dataLen + (markerSize * 2)

	if c.capacity < recSize {
		return ErrTooBig
	}

	c.hash.Reset()

	if _, err := c.hash.Write([]byte(key)); err != nil {
		return fmt.Errorf("error computing key [%s] hash: %s", key, err)
	}

	hashedKey := c.hash.Sum64()
	keyHdr, lenHdr := make([]byte, keyHdrSize), make([]byte, lenHdrSize)

	binary.LittleEndian.PutUint64(keyHdr, hashedKey)
	binary.LittleEndian.PutUint32(lenHdr, dataLen)

	c.mu.Lock()

	if c.ptr+recSize > c.capacity {
		c.ptr = 0
	}

	c.index[hashedKey] = c.ptr

	copy(c.cache[c.ptr:], keyHdr)
	c.ptr += keyHdrSize

	copy(c.cache[c.ptr:], lenHdr)
	c.ptr += lenHdrSize

	copy(c.cache[c.ptr:], marker)
	c.ptr += markerSize

	copy(c.cache[c.ptr:], value)
	c.ptr += dataLen

	copy(c.cache[c.ptr:], marker)
	c.ptr += markerSize

	c.mu.Unlock()

	return
}

func (c *bufferCache) Get(key string) ([]byte, error) {
	c.hash.Reset()

	if _, err := c.hash.Write([]byte(key)); err != nil {
		return nil, fmt.Errorf("error computing key [%s] hash: %s", key, err)
	}

	hashedKey := c.hash.Sum64()

	c.mu.RLock()

	ptr, ok := c.index[hashedKey]

	if !ok {
		c.mu.RUnlock()
		return nil, nil // no such key
	}

	storedKey := binary.LittleEndian.Uint64(c.cache[ptr:(ptr + keyHdrSize)])
	ptr += keyHdrSize

	if storedKey != hashedKey {
		delete(c.index, hashedKey)
		c.mu.RUnlock()

		return nil, ErrKeyCollision // cache key collision
	}

	dataLen := binary.LittleEndian.Uint32(c.cache[ptr:(ptr + lenHdrSize)]) + (markerSize * 2)

	if ptr+dataLen > c.capacity {
		c.mu.RUnlock()
		return nil, ErrOutOfBounds
	}

	ptr += lenHdrSize
	value := c.cache[ptr:(ptr + dataLen)]

	c.mu.RUnlock()

	if bytes.Compare(value[:markerSize], marker) != 0 || bytes.Compare(value[(dataLen - markerSize):], marker) != 0 {
		return nil, ErrCorruptedData
	}

	return value[markerSize:(dataLen - markerSize)], nil
}

func NewBuffer(capacity uint32) Cache {
	return &bufferCache{
		index:    make(map[uint64]uint32),
		cache:    make([]byte, capacity),
		capacity: capacity,
		mu:       &sync.RWMutex{},
		hash:     fnv.New64a(),
	}
}
