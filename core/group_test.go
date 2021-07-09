package core

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(k string) ([]byte, error) {
		return []byte(k), nil
	})
	k := "key"
	expect := []byte(k)
	if v, _ := f.Get(k); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback load data failed")
	}
}

var db = map[string]string{
	"Amy":   "2312",
	"Beney": "31241",
	"Roger": "9427",
}

func TestGet(t *testing.T) {
	loadCnt := make(map[string]int, len(db))
	my := NewGroup("ids", 2<<10, GetterFunc(
		func(k string) ([]byte, error) {
			log.Println("[goto DB] search key", k)
			if v, ok := db[k]; ok {
				loadCnt[k] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", k)
		}))

	for k, v := range db {
		// load from db
		if view, err := my.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %v", k)
		}

		// hit cache
		if _, err := my.Get(k); err != nil || loadCnt[k] > 1 {
			t.Fatalf("cache %s miss, but should not", k)
		}
	}

	if view, err := my.Get("unknown"); err == nil {
		t.Fatalf("should be empty, but got %s", view)
	}
}
