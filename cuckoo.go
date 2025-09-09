package cuckoofilter

import (
	"fmt"

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

func (c *CuckooFilter) Insert(data []byte) (bool, error) {
	if c.Check(data) == true {
		return false, fmt.Errorf("item already exists")
	}

	fp := Fingerprint(data, c.f)
	i1 := xxhash.Sum64(data) % c.m
	i2 := (i1 ^ xxhash.Sum64(data)) % c.m

	if c.visCount[i1] < c.b {
		c.visCount[i1]++
		c.InsertIntoBucket(i1, fp)
	} else if c.visCount[i2] < c.b {
		c.visCount[i2]++
		c.InsertIntoBucket(i2, fp)
	}

	return true, nil
}

func (c *CuckooFilter) Check(data []byte) bool {
	fp := Fingerprint(data, c.f)
	i1 := xxhash.Sum64(data) % c.m
	i2 := (i1 ^ xxhash.Sum64(data)) % c.m

	return c.CheckBucket(i1, fp) || c.CheckBucket(i2, fp)
}

func (c *CuckooFilter) Delete(data []byte) (bool, error){
	fp := Fingerprint(data, c.f)
	i1 := xxhash.Sum64(data) % c.m
	i2 := (i1 ^ xxhash.Sum64(data)) % c.m

	d, err := c.DeleteFromBucket(i1, fp)
	if err == nil {
		c.visCount[i1]--
		return d, err
	}

	d, err = c.DeleteFromBucket(i2, fp)
	if err == nil {
		c.visCount[i2]--
		return d, err
	}

	return d, err
}
