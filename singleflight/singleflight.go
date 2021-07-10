package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup // avoid reentrance
	val interface{}    // the query result
	err error
}

// Group is a singleflight shared in a cache group,
// makes sure that each key is fetched once
type Group struct {
	mu     sync.Mutex       // protect resMap
	resMap map[string]*call // store query result
}

// Do invokes the callback that binds to key, and stores
// the result as a call for sharing.
func (g *Group) Do(key string, callback func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.resMap == nil {
		g.resMap = make(map[string]*call)
	}
	if c, ok := g.resMap[key]; ok { // reuse the result
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.resMap[key] = c
	g.mu.Unlock()

	c.val, c.err = callback()
	c.wg.Done()

	g.mu.Lock()
	delete(g.resMap, key)
	g.mu.Unlock()

	return c.val, c.err
}
