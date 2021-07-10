// Package core is concurrent cache for single machine
package core

import (
	"fmt"
	"github/mycache/pb"
	"github/mycache/singleflight"
	"log"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Getter gets the value identified by key
type Getter interface {
	Get(k string) ([]byte, error) // Get key's data, from datasource
}

// GetterFunc implements Getter with a function
type GetterFunc func(k string) ([]byte, error)

func (f GetterFunc) Get(k string) ([]byte, error) {
	return f(k)
}

// Group is a cache namespace
type Group struct {
	name      string // group's name
	getter    Getter // called when all caches are missed
	mainCache cache  // cache data
	peers     PeerPicker

	// loader ensures each key is only fetched once,
	// regardless of the number of concurrent callers.
	loader *singleflight.Group
}

// Get get value from cache, if failed then get from peers,
// if failed then get from db locally
func (g *Group) Get(k string) (ByteView, error) {
	if k == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(k); ok {
		return v, nil
	}
	return g.load(k)
}

// load loads k either by sending it to a peer or
// invoking getter locally
func (g *Group) load(k string) (v ByteView, err error) {
	view, err := g.loader.Do(k, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.Pick(k); ok {
				if v, err = g.getFromPeer(peer, k); err == nil {
					return v, nil
				}
				log.Println("[MyCache] Failed to get from peer:", err)
			}
		}
		return g.getLocally(k)
	})

	if err != nil {
		return
	}
	return view.(ByteView), nil
}

// getLocally gets value identified by k from local db
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
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Fetch(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bs: res.Value}, nil
}

// NewGroup creates a group, and save to groups
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
