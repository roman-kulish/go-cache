package cache

type opFunc func(m map[string][]byte)

type channelCache struct {
	cache map[string][]byte
	op    chan opFunc
	close chan struct{}
}

func (c channelCache) Set(key string, value []byte) (err error) {
	c.op <- opFunc(func(m map[string][]byte) {
		m[key] = value
	})

	return
}

func (c channelCache) Get(key string) ([]byte, error) {
	result:= make(chan []byte)

	c.op <- opFunc(func(m map[string][]byte) {
		result <- m[key]
	})

	return <-result, nil
}

func (c channelCache) Close() {
	close(c.close)
}

func (c channelCache) runOp() {
	for {
		select {
		case op := <-c.op:
			op(c.cache)

		case <-c.close:
			return
		}
	}
}

func NewChannel(capacity uint) Cache {
	c := &channelCache{
		cache: make(map[string][]byte, capacity),
		op:    make(chan opFunc, capacity),
		close: make(chan struct{}),
	}

	go c.runOp()

	return c
}
