package cuckoofilter

import (
	"math"
	"math/rand"
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
		make([]uint64, filterSize+2),
		make([]uint, m+2),
	}
}

func (c *CuckooFilter) Insert(data []byte) bool {
	if c.Check(data) {
		return true
	}

	fp := fprint(data, c.f)
	i1 := xxhash.Sum64(data) % c.m
	i2 := (i1 ^ xxhash.Sum64([]byte{byte(fp)})) % c.m

	if c.visCount[i1] < c.b {
		c.visCount[i1]++
		c.InsertIntoBucket(i1, uint64(fp))
		return true
	}

	if c.visCount[i2] < c.b {
		c.visCount[i2]++
		c.InsertIntoBucket(i2, uint64(fp))
		return true
	}

	var i uint64
	if rand.Intn(2) == 0 {
		i = i1
	} else {
		i = i2
	}

	MaxNumKicks := int(5 * math.Log2(float64(c.n)))
	for n := 0; n < MaxNumKicks; n++ {
		evicted_fp := c.EvictFromBucket(i, uint64(fp))
		i = (i ^ xxhash.Sum64([]byte{byte(evicted_fp)})) % c.m
		fp = evicted_fp

		if c.visCount[i] < c.b {
			c.visCount[i]++
			c.InsertIntoBucket(i, fp)
			return true
		}
	}

	return false
}

func (c *CuckooFilter) Check(data []byte) bool {
	fp := fprint(data, c.f)
	i1 := xxhash.Sum64(data) % c.m
	i2 := (i1 ^ xxhash.Sum64([]byte{byte(fp)})) % c.m

	return c.CheckBucket(i1, fp) || c.CheckBucket(i2, fp)
}

func (c *CuckooFilter) Delete(data []byte) bool {
	fp := fprint(data, c.f)
	i1 := xxhash.Sum64(data) % c.m
	i2 := (i1 ^ xxhash.Sum64([]byte{byte(fp)})) % c.m

	return c.DeleteFromBucket(i1, fp) || c.DeleteFromBucket(i2, fp) || true
}
