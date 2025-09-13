# Cuckoo Filter
A Go implementation of a **Cuckoo Filter** using the Semi-Sorting Buckets technique for improved space efficiency. 

## What is a Cuckoo Filter?
A **Cuckoo Filter** is a space-efficient probabilistic data structure used to test whether an element is a member of a set—similar to a **Bloom filter**.

Key properties:

- Queries return either “possibly in set” or “definitely not in set”.

- False positives are possible, but false negatives are not.

- Supports deletion of items, which Bloom filters do not.

- For many practical workloads, Cuckoo filters can achieve lower memory overhead than optimized Bloom filters, especially when targeting moderately low false-positive rates.

## About The Implementation 
This implementation leverages the **semi-sorting buckets technique**, inspired by the paper [*Cuckoo Filter: Practically Better Than Bloom*](https://www.eecs.harvard.edu/~michaelm/postscripts/cuckoo-conext2014.pdf) to further reduce memory usage under the following conditions: 

- Each bucket contains `b = 4` entries.
- Each fingerprint is `f = 4` bits.

#### The key idea:

- The order of fingerprints inside a bucket does not affect membership queries.
- Therefore, fingerprints in each bucket can be sorted and then represented as a **compressed index** into a precomputed table of all possible sorted sequences.

#### Why This Saves Space
- Without compression, a bucket with 4 fingerprints of 4 bits each requires **16 bits**.
- With semi-sorting, there are only **3,876 unique sorted combinations**.
- These can be indexed with just **12 bits** (2¹² = 4096 > 3876).
- **Result:** 1 bit saved per fingerprint (16 → 12 bits per bucket).


## API

The Cuckoo Filter provides the following operations:

- **`Insert(item)`** Adds an item to the filter.

- **`Contain(item)`** Checks whether an item is in the filter. *Note:* This operation may return **false positives**.

- **`Delete(item)`** Removes an item from the filter.  


## Example

```go
package main

import (
    "fmt"
    cuckoo "github.com/lamasalah32/go-cuckoofilter"
)

func main() {
    // Create a new filter
    cf := cuckoo.New(100000)

    // Insert items
    cf.Insert([]byte("go"))
    cf.Insert([]byte("rust"))

    // Check for containment
    fmt.Println("Contains rust:", cf.Contain([]byte("rust")))   // true
    fmt.Println("Contains c++:", cf.Contain([]byte("c++"))) // false (may be false positive in some cases)

    // Delete an item
    cf.Delete([]byte("rust"))
    fmt.Println("Contains rust after delete:", cf.Contain([]byte("rust"))) // false
}
