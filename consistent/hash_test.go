package consistent

import (
	"strconv"
	"testing"
)

func TestHash(t *testing.T) {
	hash := New(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("2", "4", "6")

	cases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range cases {
		if hash.Locate(k) != v {
			t.Errorf("%s should map to %s", k, v)
		}
	}

	// 8, 18, 28
	hash.Add("8")

	cases["27"] = "8"

	for k, v := range cases {
		if hash.Locate(k) != v {
			t.Errorf("%s should map to %s", k, v)
		}
	}
}
