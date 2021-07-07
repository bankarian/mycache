package lru

import "testing"

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	cache := New(int64(10), nil)
	cache.Add("testKey1", String("1235"))
	if _, ok := cache.Get("testKey1"); !ok {
		t.Logf("cache hit testKey1=1235 failded")
	} else {
		t.Fatalf("cache hit testKey1=1235 succeeded, but should not")
	}
}
