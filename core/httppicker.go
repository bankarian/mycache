package core

import (
	"fmt"
	"github/mycache/consistent"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_mycache/"
	defaultReplicas = 50
)

// HTTPPicker implements PeerPicker for a pool of HTTP peers.
type HTTPPicker struct {
	self string
	// prefix of the communication address between nodes,
	// http://xx.com/_mycache/ serves as the default prefix.
	basePath    string
	mu          sync.Mutex
	peers       *consistent.Map
	httpGetters map[string]*httpGetter // get key by url, eg. "http://localhost:8080"
}

func NewHTTPPicker(self string) *HTTPPicker {
	return &HTTPPicker{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPicker) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s",
		p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPicker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPicker serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// /<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request, missing parts, should be /basepath/groupname/key", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := w.Write(view.Slice()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Set sets the pool's list of peers, discards the old ones
func (p *HTTPPicker) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistent.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPicker) Pick(key string) (Peer, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Println("self=", p.self)
	if peer := p.peers.Locate(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPicker)(nil)
