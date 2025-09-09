package cuckoofilter

import (
    "bytes"
    "math"
)

func swap[T any](s []T, i, j uint) {
    s[i], s[j] = s[j], s[i]
}

func ExtractBits(x uint64, start, end uint) uint64 {
	lsbStart := 63 - end
	lsbEnd := 63 - start

	width := lsbEnd - lsbStart + 1
	mask := (uint64(1) << width) - 1

	return (x >> lsbStart) & mask
}

func IndexLockup(fp []byte) uint {
	l, r := 0, len(Comb)-1
	for l <= r {
		mid := l + (r-l)/2

		if bytes.Compare(Comb[mid], fp) < 0 {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}

	return uint(l)
}

func MinFingerprintBits(n uint64, b uint) uint {
	if n == 0 || b == 0 {
		return 0
	}

	return uint(math.Max(4, math.Ceil(math.Log2(float64(n))/float64(b))))
}

