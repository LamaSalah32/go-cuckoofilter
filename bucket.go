package cuckoofilter

import (
	"math/rand"
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
	startBit := idx * c.bucketSize
	bits := c.getBitsFromBucket(startBit, startBit + c.bucketSize - 1)
	first12 := bits >> (c.bucketSize - 12)

	combination := append([]byte(nil), Comb[first12]...)

	insertPos := c.b - c.visCount[idx]
	combination[insertPos] = byte(fp >> (c.f - 4))
	c.setBitsInBucket(
		fp & ((1 << (c.f - 4)) - 1), 
		startBit + 12 + uint64(insertPos)*uint64(c.f-4), 
		startBit + 12 + uint64(insertPos+1)*uint64(c.f-4)-1,
	)

	resBits := c.getBitsFromBucket(startBit+12, startBit+c.bucketSize-1)
	combination, resBits = Sort(combination, resBits, c.b, c.f)

	newFirst12 := uint64(IndexLockup(combination))
	c.setBitsInBucket(newFirst12, startBit, startBit+11)
	c.setBitsInBucket(resBits, startBit+12, startBit+c.bucketSize-1)

	return true
}

func (c *CuckooFilter) ReplaceInBucket(idx uint64, pos int, fp uint64) {
    startBit := idx * c.bucketSize
    first12 := c.getBitsFromBucket(startBit, startBit+11)

    combination := append([]byte(nil), Comb[first12]...)
    combination[pos] = byte(fp >> (c.f - 4))
	c.setBitsInBucket(
		fp & ((1 << (c.f - 4)) - 1),
		startBit + 12 + uint64(pos)*uint64(c.f-4),
		startBit + 12 + uint64(pos+1)*uint64(c.f-4)-1,
	)

    resBits := c.getBitsFromBucket(startBit+12, startBit+c.bucketSize-1)
    combination, resBits = Sort(combination, resBits, c.b, c.f)

	newFirst12 := uint64(IndexLockup(combination))
	c.setBitsInBucket(newFirst12, startBit, startBit+11)
    c.setBitsInBucket(resBits, startBit+12, startBit+c.bucketSize-1)
}

func (c *CuckooFilter) CheckBucket(idx uint64, fp uint64) bool {
	startBit := idx * c.bucketSize
	bits := c.getBitsFromBucket(startBit, startBit + c.bucketSize - 1)
	first12 := bits >> (c.bucketSize - 12)

	combination := append([]byte(nil), Comb[first12]...)
	last4 := fp >> (c.f - 4)

	for i := uint64(0); i < uint64(c.b); i++ {
		if combination[i] == byte(last4) {
			restBits := c.getBitsFromBucket(
				uint64(startBit + 12 + i*uint64(c.f-4)),
				uint64(startBit + 12 + (i+1)*uint64(c.f-4) - 1),
			)

			fpRest := fp & ((1 << (c.f - 4)) - 1)
			if restBits == fpRest {
				return true
			}
		}
	}

	return false
}

func (c *CuckooFilter) DeleteFromBucket(idx uint64, fp uint64) bool {
	startBit := idx * c.bucketSize
	bits := c.getBitsFromBucket(startBit, startBit + c.bucketSize - 1)
	first12 := bits >> (c.bucketSize - 12)

	combination := append([]byte(nil), Comb[first12]...)
	last4 := fp >> (c.f - 4)
	
	for i := uint64(0); i < uint64(c.b); i++ {
		if combination[i] == byte(last4) {
			// check rest
			sRest := uint64(startBit + 12 + uint64(i*uint64(c.f-4)))
			restBits := c.getBitsFromBucket(sRest, sRest+uint64(c.f-4)-1)

			fpRest := fp & ((1 << (c.f - 4)) - 1)
			
			if restBits == fpRest {
				// delete
				combination[i] = 0
				c.setBitsInBucket(0, sRest, sRest+uint64(c.f-4)-1)
				rest := c.getBitsFromBucket(startBit+12, startBit+c.bucketSize-1)
				
				combination, rest = Sort(combination, rest, c.b, c.f)

				newFirst12 := uint64(IndexLockup(combination))
				c.setBitsInBucket(newFirst12, startBit, startBit+11)
				c.setBitsInBucket(rest, startBit+12, startBit+c.bucketSize-1)

				c.visCount[idx]--
				return true
			}
		}
	}

	return false
}

func (c *CuckooFilter) EvictFromBucket(idx uint64, fp uint64) uint64 {
	r := rand.Intn(int(c.b))
	startBit := idx * c.bucketSize
	var first12 = c.getBitsFromBucket(startBit, startBit+11)
	pos_r := Comb[first12][r]

	restBits := c.getBitsFromBucket(
		startBit+12+uint64(r)*uint64(c.f-4),
		startBit+12+uint64(r+1)*uint64(c.f-4)-1,
	)

	evicted_fp := (uint64(pos_r) << (c.f - 4)) | restBits
	c.ReplaceInBucket(idx, r, fp)
	return evicted_fp
}