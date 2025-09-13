package cuckoofilter

import (
	"math/bits"
	"bytes"
	"math"
)

var Comb = GenerateCombinations(16, 4)

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

func nextPow2(m uint64) uint64 {
    if m == 0 {
        return 1
    }
	
    if (m & (m - 1)) == 0 {
        return m 
    }

    return 1 << (64 - bits.LeadingZeros64(m-1))
}

func Sort(combIdx []byte, rest uint64, b uint, f uint) ([]byte, uint64) {
	slotBits := f - 4
	bits := make([]uint64, b)

	for i := uint(0); i < b; i++ {
		shift := (b - 1 - i) * slotBits  
		mask := (uint64(1) << slotBits) - 1
		bits[i] = (rest >> shift) & mask
	}

	for i := uint(0); i < b; i++ {
		for j := uint(0); j < b-i-1; j++ {
            if combIdx[j] > combIdx[j+1] {
                combIdx[j], combIdx[j+1] = combIdx[j+1], combIdx[j]
                bits[j], bits[j+1] = bits[j+1], bits[j]
			}
        }
    }

	var res uint64 = 0
	for i := uint(0); i < b; i++ {
		res |= bits[i] << ((b - 1 - i) * slotBits)
	}

	return combIdx, res
}