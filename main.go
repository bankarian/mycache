package main

import (
	"flag"
	"fmt"
	"github/mycache/core"
	"log"
	"net/http"
)

var scoreDB = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *core.Group {
	return core.NewGroup("scores", 2<<10, core.GetterFunc(
		func(k string) ([]byte, error) {
			log.Println("[Query DB] search key", k)
			if v, ok := scoreDB[k]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", k)
		}))
}

// start cache server
func startCache(addr string, addrs []string, myc *core.Group) {
	peers := core.NewHTTPPicker(addr)
	peers.Set(addrs...)
	myc.RegisterPeers(peers)
	log.Println("mycache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[len("http://"):], peers))
}

// start api server for user
func startAPI(apiAddr string, myc *core.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := myc.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.Slice())
		}))
	log.Println("frontend server runing at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[len("http://"):], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8081, "MyCache server port")
	flag.BoolVar(&api, "api", false, "Start an api sever?")
	flag.Parse()

	apiAddr := "http://localhost:6789"
	addrMap := map[int]string{
		8081: "http://localhost:8081",
		8082: "http://localhost:8082",
		8083: "http://localhost:8083",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	myc := createGroup()
	if api {
		go startAPI(apiAddr, myc)
	}
	startCache(addrMap[port], addrs, myc)
}
