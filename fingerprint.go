package cuckoofilter

import (
	"github.com/cespare/xxhash/v2"
)

func fprint(item []byte, fpSize uint) uint64 {
	h := xxhash.Sum64(item)
	if fpSize > 0 || fpSize < 64 {
		mask := uint64(1<<fpSize) - 1
		h &= mask
	} 

	if h == 0 {
		h = 1
	}
	return h
}

func hash (item []byte)  uint64 {
	return  xxhash.Sum64(item)
}
