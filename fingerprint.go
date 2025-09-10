package cuckoofilter

import (
	"github.com/cespare/xxhash/v2"
)

func fprint(item []byte, fpSize uint) uint64 {
	h := xxhash.Sum64(item)
	if fpSize < 64 {
		mask := (uint64(1) << fpSize) - 1
		h &= mask
	}
	return uint64(h)
}