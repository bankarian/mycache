package consistant

import (
	"hash/crc32"
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

// New .
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

// Add given keys, construct real nodes.
// Each node has replicas virtual nodes
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

// Locate gets the closest node's key
func (m *Map) Locate(k string) string {
	if len(k) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(k)))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.dict[m.keys[idx%len(m.keys)]]
}
