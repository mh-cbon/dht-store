package dhtstore

import (
	"crypto/sha1"
	"fmt"
)

func hashSha1(s ...string) string {
	h := sha1.New()
	for _, v := range s {
		h.Write([]byte(v))
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
