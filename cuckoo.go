package cuckoofilter

import (
	"math"
	"math/rand"
)

type CuckooFilter struct {
	n          uint64 // number of items in the filter
	b          uint   // number of slots per bucket
	f          uint   // number of fingerprint bits
	m          uint64 // number of buckets
	mask       uint64
	bucketSize uint64
	bucket     []uint64
	visCount   []uint
}

func New(n uint64) *CuckooFilter {
	b := uint(4)
	f := uint(MinFingerprintBits(n, b))
	bucketSize := uint64((12+((f/4-1)*12)+((f%4)*4))*(b/4) + ((b % 4) * f))
	filterSize := uint64((n*uint64(f) + 63) / 64)
	m := nextPow2((filterSize * 64) / bucketSize)
	filterSize = ((m * bucketSize) + 63) / 64
	mask := uint64(m - 1)

	return &CuckooFilter{
		n,
		b,
		f,
		m,
		mask,
		bucketSize,
		make([]uint64, filterSize+2),
		make([]uint, m),
	}
}

func (c *CuckooFilter) Insert(data []byte) bool {
	if c.Contain(data) {
		return true
	}

	fp := fprint(data, c.f)
	i1 := hash(data) & c.mask
	hfp := hash([]byte{byte(fp)}) & c.mask
	i2 := (i1 ^ hfp) & c.mask

	if c.visCount[i1] < c.b {
		c.visCount[i1]++
		c.InsertIntoBucket(i1, int(c.b - c.visCount[i1]), fp)
		return true	
	}

	if c.visCount[i2] < c.b {
		c.visCount[i2]++
		c.InsertIntoBucket(i2, int(c.b - c.visCount[i2]), fp)
		return true
	}

	i := i1
    if rand.Intn(2) == 1 {
        i = i2
    }

	MaxNumKicks := int(5 * math.Log2(float64(c.n)))
	for n := 0; n < MaxNumKicks; n++ {
		fp := c.EvictFromBucket(i, fp)
		hfp := hash([]byte{byte(fp)}) & c.mask
		i = (i ^ hfp) & c.mask

		if c.visCount[i] < c.b {
			c.visCount[i]++
			c.InsertIntoBucket(i, int(c.b - c.visCount[i]), fp)
			return true
		}
	}

	return false
}

func (c *CuckooFilter) Contain(data []byte) bool {
	fp := fprint(data, c.f)
	i1 := hash(data) & c.mask
	hfp := hash([]byte{byte(fp)}) & c.mask
	i2 := (i1 ^ hfp) & c.mask


	return c.ContainBucket(i1, fp) || c.ContainBucket(i2, fp)
}

func (c *CuckooFilter) Delete(data []byte) bool {
	fp := fprint(data, c.f)
	i1 := hash(data) & c.mask
	hfp := hash([]byte{byte(fp)}) & c.mask
	i2 := (i1 ^ hfp) & c.mask

	return c.DeleteFromBucket(i1, fp) || c.DeleteFromBucket(i2, fp) || true
}