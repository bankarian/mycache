package mycache

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestHTTPPool(t *testing.T) {
	NewGroup("ids", 2<<10, GetterFunc(
		func(k string) ([]byte, error) {
			log.Println("[goto DB] search key", k)
			if v, ok := db[k]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", k)
		}))

	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	t.Log("mycache is running at", addr)

	http.ListenAndServe(addr, peers)
}
