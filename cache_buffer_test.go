package cache

import (
	"bytes"
	"testing"
)

func TestCacheBuffer(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
	}{
		{
			name:  "a",
			value: []byte("hello"),
		},
		{
			name:  "b",
			value: []byte("world"),
		},
	}

	c := NewBuffer(64)

	for _, test := range tests {
		if err := c.Set(test.name, test.value); err != nil {
			t.Fatal(err)
		}
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := c.Get(test.name)

			if err != nil {
				t.Error(err)
			} else if v == nil || bytes.Compare(test.value, v) != 0 {
				t.Errorf("value mismatch for key %s", test.name)
			}
		})
	}
}
