package cuckoofilter

import (
	"github.com/cespare/xxhash/v2"
)

type fingerprint []byte

func Fingerprint(data []byte, f uint) uint64 {
	h := xxhash.Sum64(data)

	if f >= 64 {
		return h
	}

	mask := uint64((1 << f) - 1)
	fp := h & mask
	return uint64(fp)
}
