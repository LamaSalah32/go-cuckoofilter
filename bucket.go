package cuckoofilter

import (
	"math/rand"
)

func (c *CuckooFilter) getBitsFromBucket(start, end uint64) (bucketBits uint64) {
	if start > end{
		return 0
	}

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
	if start > end{
		return
	}

	if end/64 == start/64 {
		bitLen := end - start + 1
		mask := ((uint64(1) << bitLen) - 1) << (64 - (end%64 + 1))
		c.bucket[start/64] = (c.bucket[start/64] & ^mask) | (uint64(newBits) << (64 - (end%64 + 1)))
	} else {
		leftBits := 64 - (start % 64)
		rightBits := end%64 + 1

		leftPart := newBits >> rightBits
		rightPart := newBits & ((1 << rightBits) - 1)

		c.setBitsInBucket(leftPart, start, start+uint64(leftBits)-1)
		c.setBitsInBucket(rightPart, end-uint64(rightBits)+1, end)
	}
}

func (c *CuckooFilter) InsertIntoBucket(idx uint64, pos int, fp uint64) bool {
	startBit := idx * c.bucketSize
	bits := c.getBitsFromBucket(startBit, startBit + c.bucketSize - 1)
	combIdx := bits >> (c.bucketSize - 12)

	combination := append([]byte(nil), Comb[int(combIdx)]...)
	combination[pos] = byte(fp >> (c.f - 4))

	c.setBitsInBucket(
		fp & ((1 << (c.f - 4)) - 1), 
		startBit + 12 + uint64(pos)*uint64(c.f-4), 
		startBit + 12 + uint64(pos+1)*uint64(c.f-4)-1,
	)

	resBits := c.getBitsFromBucket(startBit+12, startBit+c.bucketSize-1)
	combination, resBits = Sort(combination, resBits, c.b, c.f)

	newcombIdx := uint64(IndexLockup(combination))
	c.setBitsInBucket(newcombIdx, startBit, startBit+11)
	c.setBitsInBucket(resBits, startBit+12, startBit+c.bucketSize-1)

	return true
}

func (c *CuckooFilter) CheckBucket(idx uint64, fp uint64) bool {
	startBit := idx * c.bucketSize
	bits := c.getBitsFromBucket(startBit, startBit + c.bucketSize - 1)
	combIdx := bits >> (c.bucketSize - 12)

	combination := append([]byte(nil), Comb[int(combIdx)]...)

	for i := uint64(0); i < uint64(c.b); i++ {
		restBits := c.getBitsFromBucket(
			uint64(startBit + 12 + i*uint64(c.f-4)),
			uint64(startBit + 12 + (i+1)*uint64(c.f-4) - 1),
		)

		val :=  uint64(combination[i] << (c.f - 4)) | restBits
		if val == fp{
			return true
		}

	}

	return false
}

func (c *CuckooFilter) DeleteFromBucket(idx uint64, fp uint64) bool {
	startBit := idx * c.bucketSize
	bits := c.getBitsFromBucket(startBit, startBit + c.bucketSize - 1)
	combIdx := bits >> (c.bucketSize - 12)

	combination := append([]byte(nil), Comb[int(combIdx)]...)
	
	for i := uint64(0); i < uint64(c.b); i++ {
		restBits := c.getBitsFromBucket(
			uint64(startBit + 12 + i*uint64(c.f-4)),
			uint64(startBit + 12 + (i+1)*uint64(c.f-4) - 1),
		)

		val :=  uint64(combination[i] << (c.f - 4)) | restBits

		if val == fp {
			sRest := uint64(startBit + 12 + uint64(i*uint64(c.f-4)))
		
			combination[i] = 0
			c.setBitsInBucket(0, sRest, sRest+uint64(c.f-4)-1)
			rest := c.getBitsFromBucket(startBit+12, startBit+c.bucketSize-1)
			
			combination, rest = Sort(combination, rest, c.b, c.f)

			newcombIdx := uint64(IndexLockup(combination))
			c.setBitsInBucket(newcombIdx, startBit, startBit+11)
			c.setBitsInBucket(rest, startBit+12, startBit+c.bucketSize-1)

			c.visCount[idx]--
			return true
			
		}
	}

	return false
}

func (c *CuckooFilter) EvictFromBucket(idx uint64, fp uint64) uint64 {
	r := rand.Intn(int(c.b))
	startBit := idx * c.bucketSize
	combIdx := c.getBitsFromBucket(startBit, startBit+11)

	val_r := Comb[combIdx][r]
	restBits := c.getBitsFromBucket(
		startBit+12+uint64(r)*uint64(c.f-4),
		startBit+12+uint64(r+1)*uint64(c.f-4)-1,
	)

	evicted_fp := (uint64(val_r) << (c.f - 4)) | restBits

	c.InsertIntoBucket(idx, r, fp)
	return evicted_fp
}
