package cuckoofilter

import (
	"fmt"
)

func (c *CuckooFilter) getBitsFromBucket(start, end uint64) (bucketBits uint64) {
	if end/64 == start/64 {
		bucketBits = ExtractBits(c.bucket[start/64], uint(start%64), uint(end%64))
	} else {
		part1 := ExtractBits(c.bucket[start/64], uint(start%64), 63)
		part2 := ExtractBits(c.bucket[end/64], 0, uint(end%64))
		bucketBits = part1<<(end%64+1) | part2
	}

	return
}

func (c *CuckooFilter) setBitsInBucket(newBits uint64, start, end uint64) {
	if end/64 == start/64 {
		bitLen := end - start + 1
		mask := ((uint64(1) << bitLen) - 1) << (64 - (end%64 + 1))
		c.bucket[start/64] = (c.bucket[start/64] & ^mask) | (uint64(newBits) << (64 - (end%64 + 1)))
	} else {
		leftBits := 64 - (start % 64)
		rightBits := end%64 + 1

		leftPart := newBits >> rightBits
		rightPart := newBits & ((1 << rightBits) - 1)

		leftMask := ((uint64(1) << leftBits) - 1) << (64 - leftBits)
		c.bucket[start/64] = (c.bucket[start/64] & ^leftMask) | (uint64(leftPart) << (64 - leftBits))

		rightMask := (uint64(1) << rightBits) - 1<<(64-rightBits)
		c.bucket[end/64] = (c.bucket[end/64] & ^rightMask) | (uint64(rightPart) << (64 - rightBits))
	}
}

func (c *CuckooFilter) InsertIntoBucket(idx, fp uint64) bool {
	var startBit uint64
	if idx == 0 {
		startBit = 0
	} else {
		startBit = (idx - 1) * c.bucketSize
	}
	endBit := startBit + c.bucketSize - 1
	var bits = c.getBitsFromBucket(startBit, endBit)

	first12Bits := bits >> (c.bucketSize - 12)

	combIdx := append([]byte(nil), Comb[first12Bits]...)

	lastFourBits := fp >> (c.f - 4)
	combIdx[0] = byte(lastFourBits)

	var correctPos uint
	for i := uint(0); i < 3; i++ {
		if combIdx[i] > combIdx[i+1] {
			swap(combIdx, i, i+1)
			correctPos = i + 1
		}
	}

	newBits := uint64(IndexLockup(combIdx))

	c.setBitsInBucket(newBits, startBit, startBit+11)

	restBits := fp & ((1 << (c.f - 4)) - 1)

	sRest := startBit + 12 + uint64(correctPos*(c.f-4))
	c.setBitsInBucket(restBits, sRest, sRest+uint64(c.f-4)-1)

	return true
}

func (c *CuckooFilter) CheckBucket(idx uint64, fp uint64) bool {
	var startBit uint64
	if idx == 0 {
		startBit = 0
	} else {
		startBit = (idx - 1) * c.bucketSize
	}

	endBit := startBit + c.bucketSize - 1
	var bits = c.getBitsFromBucket(startBit, endBit)

	first12Bits := bits >> (c.bucketSize - 12)

	combIdx := append([]byte(nil), Comb[first12Bits]...)
	lastFourBits := fp >> (c.f - 4)

	for i := uint64(0); i < uint64(c.b); i++ {
		if combIdx[i] == byte(lastFourBits) {
			// check rest
			sRest := uint64(startBit + 12 + uint64(i*uint64(c.f-4)))
			restBits := c.getBitsFromBucket(sRest, sRest+uint64(c.f-4)-1)

			fpRest := fp & ((1 << (c.f - 4)) - 1)

			if restBits == fpRest {
				return true
			}
		}
	}

	return false
}

func (c *CuckooFilter) DeleteFromBucket(idx uint64, fp uint64) (bool, error) {
	var startBit uint64
	if idx == 0 {
		startBit = 0
	} else {
		startBit = (idx - 1) * c.bucketSize
	}

	endBit := startBit + c.bucketSize - 1
	var bits = c.getBitsFromBucket(startBit, endBit)


	first12Bits := bits >> (c.bucketSize - 12)
	combIdx := append([]byte(nil), Comb[first12Bits]...)
	lastFourBits := fp >> (c.f - 4)

	for i := uint64(0); i < uint64(c.b); i++ {
		if combIdx[i] == byte(lastFourBits) {
			// check rest
			sRest := uint64(startBit + 12 + uint64(i*uint64(c.f-4)))
			restBits := c.getBitsFromBucket(sRest, sRest+uint64(c.f-4)-1)

			fpRest := fp & ((1 << (c.f - 4)) - 1)

			if restBits == fpRest {
				// delete
				combIdx[i] = 0
				for j := uint(i); j > 0; j-- {
					swap(combIdx, j, uint(j-1))
				}

				newBits := uint64(IndexLockup(combIdx))
				c.setBitsInBucket(newBits, startBit, startBit+11)
				c.setBitsInBucket(0, sRest, sRest+uint64(c.f-4)-1)
				return true, nil
			}
		}
	}

	return false, fmt.Errorf("can not find fingerprint in bucket to delete")
}