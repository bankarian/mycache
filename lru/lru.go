package lru

import "container/list"

// Cache is a LRU locate. Not safe for concurrency.
type Cache struct {
	maxBytes int64      // max usable bytes
	bytesCnt int64      // how many bytes are used
	lru      *list.List // head is the move active element
	locate   map[string]*list.Element
	onEvict  func(k string, v Value)
}

// Value should has Len()
type Value interface {
	Len() int // how many bytes are used
}

// entry is lru list's element
type entry struct {
	k string
	v Value
}

// New constructor of Cache
func New(maxBytes int64, onEvict func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		bytesCnt: 0,
		lru:      list.New(),
		locate:   make(map[string]*list.Element),
		onEvict:  onEvict,
	}
}

func (c *Cache) Get(k string) (v Value, ok bool) {
	if elm, ok := c.locate[k]; ok {
		c.lru.MoveToFront(elm)
		kv := elm.Value.(*entry)
		return kv.v, true
	}
	return
}

func (c *Cache) Remove() {
	if elm := c.lru.Back(); elm != nil {
		c.lru.Remove(elm)
		kv := elm.Value.(*entry)
		delete(c.locate, kv.k)
		c.bytesCnt -= int64(len(kv.k)) + int64(kv.v.Len())
		if c.onEvict != nil {
			c.onEvict(kv.k, kv.v)
		}
	}

}

// Add adds new value to the cache,
// replace the old value if key exists
func (c *Cache) Add(k string, v Value) {
	if elm, ok := c.locate[k]; ok {
		c.lru.MoveToFront(elm)
		kv := elm.Value.(*entry)
		c.bytesCnt += int64(v.Len()) - int64(kv.v.Len())
		kv.v = v
	} else {
		elm := c.lru.PushFront(&entry{k, v})
		c.locate[k] = elm
		c.bytesCnt += int64(len(k)) + int64(v.Len())
	}
	// should not oversize
	for c.maxBytes != 0 && c.bytesCnt > c.maxBytes {
		c.Remove()
	}
}

func (c *Cache) Len() int {
	return c.lru.Len()
}
