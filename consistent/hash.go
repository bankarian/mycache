package consistent

import (
	"hash/crc32"
	"log"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

var defaultHash = crc32.ChecksumIEEE

// Map contains all hashed keys
type Map struct {
	hash     Hash
	replicas int            // how many replicas (virtual nodes)
	keys     []int          // sorted, store all keys
	dict     map[int]string // virtual key to real key
}

// New construct a consistent hash, fn can be nil
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		dict:     make(map[int]string),
	}
	if m.hash == nil {
		m.hash = defaultHash
	}
	return m
}

// Add given keys.
// Each key has replicas virtual key
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			virtualKey := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, virtualKey)
			m.dict[virtualKey] = key
		}
	}
	sort.Ints(m.keys)
}

// Locate gets the closest node's key, return "" if not found
func (m *Map) Locate(k string) string {
	log.Println("[locate] k=", k)
	if len(k) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(k)))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.dict[m.keys[idx%len(m.keys)]]
}
