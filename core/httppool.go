package core

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github/mycache/consistent"
	"github/mycache/pb"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_mycache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	self string
	// prefix of the communication address between nodes,
	// http://xx.com/_mycache/ serves as the default prefix.
	basePath     string
	mu           sync.Mutex
	peers        *consistent.Map
	httpFetchers map[string]*httpFetcher // get key by url, eg. "http://localhost:8080"
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s",
		p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
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

	byts, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(byts)
}

// Set sets the pool's list of peers(url), discards the old ones
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistent.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpFetchers = make(map[string]*httpFetcher, len(peers))
	for _, peer := range peers {
		p.httpFetchers[peer] = &httpFetcher{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) Pick(key string) (Peer, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Locate(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpFetchers[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
