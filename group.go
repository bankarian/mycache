// Package mycache is concurrent cache for single machine
package mycache

import (
	"fmt"
	"log"
	"sync"
)

// Getter loads data for key, which is called
// when cache is missed
type Getter interface {
	Get(k string) ([]byte, error) // Get key's data, from datasource
}

// GetterFunc is implementation of Getter
type GetterFunc func(k string) ([]byte, error)

func (f GetterFunc) Get(k string) ([]byte, error) {
	return f(k)
}

// Group is a cache namespace
type Group struct {
	name      string // group's name
	getter    Getter // called when cached miss
	mainCache cache  // cache data
}

func (g *Group) Get(k string) (ByteView, error) {
	if k == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(k); ok {
		log.Println("[MyCache] hit")
		return v, nil
	}
	return g.load(k)
}

func (g *Group) load(k string) (ByteView, error) {
	return g.getLocally(k)
}

func (g *Group) getLocally(k string) (ByteView, error) {
	byts, err := g.getter.Get(k)
	if err != nil {
		return ByteView{}, err
	}
	v := ByteView{bs: clone(byts)}

	g.mainCache.add(k, v)
	return v, nil
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, maxBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil error")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{maxBytes: maxBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group, or
// nil if there's no such group
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}
