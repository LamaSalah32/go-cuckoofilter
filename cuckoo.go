package cuckoofilter

import (
	"bytes"
	"fmt"
	"math"
	
	"github.com/cespare/xxhash/v2"
)

var Comb = GenerateCombinations(16, 4)

type CuckooFilter struct {
	n          uint64 // number of items in the filter
	b          uint   // number of slots per bucket
	f          uint   // number of fingerprint bits
	m          uint64 // number of buckets
	bucketSize uint64
	bucket     []uint64
	visCount   []uint
}

func New(n uint64) *CuckooFilter {
	b := uint(4)
	f := uint(MinFingerprintBits(n, b))
	bucketSize := uint64((12+((f/4-1)*12)+((f%4)*4))*(b/4) + ((b % 4) * f))
	filterSize := uint64((n*uint64(f) + 63) / 64)
	m := (filterSize + 1) / bucketSize

	return &CuckooFilter{
		n,
		b,
		f,
		m,
		bucketSize,
		make([]uint64, filterSize),
		make([]uint, m),
	}
}

func ExtractBits(x uint64, start, end uint) uint64 {
	lsbStart := 63 - end
	lsbEnd := 63 - start

	width := lsbEnd - lsbStart + 1
	mask := (uint64(1) << width) - 1

	return (x >> lsbStart) & mask
}

func (c *CuckooFilter) getBitsFromBucket(start, end uint64) (bucketBits uint64) {
	fmt.Println(start/64, end/64, start%64, end%64)

	if end/64 == start/64 {
		bucketBits = ExtractBits(c.bucket[start/64], uint(start%64), uint(end%64))
		fmt.Printf("bits = %b\n", bucketBits)
	} else {
		part1 := ExtractBits(c.bucket[start/64], uint(start%64), 63)
		part2 := ExtractBits(c.bucket[end/64], 0, uint(end%64))
		fmt.Printf("part1 = %b\n", part1)
		fmt.Printf("part2 = %b\n", part2)

		bucketBits := part1<<(end%64+1) | part2
		fmt.Printf("combined bits = %b\n", bucketBits)
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
	startBit := (idx - 1) * c.bucketSize
	endBit := startBit + c.bucketSize - 1
	var bits = c.getBitsFromBucket(startBit, endBit)

	fmt.Printf("Original bits = %b\n", bits)

	first12Bits := bits >> (c.bucketSize - 12)
	fmt.Printf("First 12 bits = %b\n", first12Bits)

	combIdx := append([]byte(nil), Comb[first12Bits]...)
	fmt.Println(combIdx)

	lastFourBits := fp >> (c.f - 4)
	fmt.Println(lastFourBits)
	combIdx[0] = byte(lastFourBits)
	fmt.Println(combIdx)

	var correctPos uint 
	for i := uint(0); i < 3; i++ {
		if (combIdx[i] > combIdx[i+1]) {
			swap(combIdx, i, i+1)
			fmt.Println("Swapped:", combIdx)
			correctPos = i+1
		}
	}

	fmt.Println("Correct Position:", correctPos)

	fmt.Println(combIdx)
	newBits := uint64(IndexLockup(combIdx))

	c.setBitsInBucket(newBits, startBit, startBit + 11)
	
	restBits := fp & ((1 << (c.f - 4)) - 1) 
	fmt.Println("restBits =", restBits)
	fmt.Println("restBits start =",  12 + uint64(correctPos * (c.f-4)))
	c.setBitsInBucket(restBits, startBit + 12 + uint64(correctPos * (c.f-4)), endBit)

	return true
}

func (c *CuckooFilter) Insert(data []byte) {
	fmt.Println(c.b, c.f, c.m, c.bucketSize)
	fmt.Println("***************")

	fp := Fingerprint(data, c.f)
	i1 := xxhash.Sum64(data) % c.m
	i2 := (i1 ^ xxhash.Sum64(data)) % c.m

	fmt.Println(c.visCount[i1], i1, i2)

	if c.visCount[i1] < c.b {
		c.visCount[i1]++
		c.InsertIntoBucket(i1, fp)
	}

	fmt.Println("After :", c.visCount[i1], fp, i1, i2)
	fmt.Println("------------------")
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

