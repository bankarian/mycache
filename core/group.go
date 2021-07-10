// Package core is concurrent cache for single machine
package core

import (
	"fmt"
	"github/mycache/singleflight"
	"log"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
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
	peers     PeerPicker
	loader    *singleflight.Group
}

// Get get value from cache, if failed then get from peers,
// if failed then get from db locally
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

// load get from peers first, if failed, then go for local db
func (g *Group) load(k string) (v ByteView, err error) {
	view, err := g.loader.Do(k, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.Pick(k); ok {
				if v, err = g.getFromPeer(peer, k); err == nil {
					return v, nil
				}
				log.Println("[MyCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(k)
	})

	if err != nil {
		return
	}
	return view.(ByteView), nil
}

// getLocally gets data from local db
func (g *Group) getLocally(k string) (ByteView, error) {
	byts, err := g.getter.Get(k)
	if err != nil {
		return ByteView{}, err
	}
	v := ByteView{bs: clone(byts)}

	g.mainCache.add(k, v)
	return v, nil
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("[RegisterPeers] called more than once")
	}
	g.peers = peers
}

// getFromPeer .
func (g *Group) getFromPeer(peer Peer, key string) (ByteView, error) {
	byts, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bs: byts}, nil
}

// NewGroup constructs a group, and save to groups
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
		loader:    &singleflight.Group{},
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
