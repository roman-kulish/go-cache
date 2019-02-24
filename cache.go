package cache

type Cache interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
}
