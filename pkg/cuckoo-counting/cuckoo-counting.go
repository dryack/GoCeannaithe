package cuckoo_counting

import (
	"github.com/dryack/GoCeannaithe/pkg/common"
)

const (
	maxKicks = 500
	maxCount = 255 // Maximum count value (8-bit counter)
)

// Bucket represents a bucket in the Counting Cuckoo Filter
type Bucket[T common.Hashable] struct {
	keys   [4]T
	counts [4]uint8
	empty  T
}

// CountingCuckooFilter represents a Counting Cuckoo Filter
type CountingCuckooFilter[T common.Hashable] struct {
	buckets      []*Bucket[T]
	numBuckets   uint32
	count        uint32
	hashFunction func(T, uint32) (uint64, error)
	hashEnum     uint8
}

// NewCountingCuckooFilter creates a new Counting Cuckoo Filter with the specified number of buckets
func NewCountingCuckooFilter[T common.Hashable](numBuckets uint32) *CountingCuckooFilter[T] {
	buckets := make([]*Bucket[T], numBuckets)
	for i := range buckets {
		buckets[i] = &Bucket[T]{empty: common.ZeroValue[T]()}
	}
	cf := &CountingCuckooFilter[T]{
		buckets:    buckets,
		numBuckets: numBuckets,
		count:      0,
	}
	// cf.WithHashFunction(hashFunc)  TODO:  move this over to WithAutoConfigure() when the time comes
	return cf
}

// WithHashFunction sets the hash function to be used by the Counting Cuckoo Filter
func (cf *CountingCuckooFilter[T]) WithHashFunction(hashFunc uint8) *CountingCuckooFilter[T] {
	switch hashFunc {
	case common.Murmur3:
		cf.hashFunction = common.HashKeyMurmur3[T]
	case common.Sha256:
		cf.hashFunction = common.HashKeySha256[T]
	case common.Sha512:
		cf.hashFunction = common.HashKeySha512[T]
	case common.SipHash:
		cf.hashFunction = common.HashKeySipHash[T]
	case common.XXhash:
		cf.hashFunction = common.HashKeyXXhash[T]
	default:
		cf.hashFunction = common.HashKeyMurmur3[T]
	}
	cf.hashEnum = hashFunc
	return cf
}

// Insert inserts a key into the Counting Cuckoo Filter and increments its count
func (cf *CountingCuckooFilter[T]) Insert(key T) bool {
	i1, i2, fp := cf.getIndicesAndFingerprint(key, 0)

	if cf.increment(key, i1) || cf.increment(key, i2) {
		return true
	}

	i := i1
	for k := 0; k < maxKicks; k++ {
		f := cf.buckets[i].swapFingerprint(fp)
		fp = f

		i = cf.getAlternateIndex(fp, i, 0)
		if cf.insert(key, i) {
			return true
		}
	}

	return false
}

// Lookup checks if a key exists in the Counting Cuckoo Filter and returns its count
func (cf *CountingCuckooFilter[T]) Lookup(key T) uint8 {
	i1, i2, _ := cf.getIndicesAndFingerprint(key, 0)
	count := cf.buckets[i1].getCount(key)
	if count == 0 {
		count = cf.buckets[i2].getCount(key)
	}
	return count
}

// Delete removes a key from the Counting Cuckoo Filter
func (cf *CountingCuckooFilter[T]) Delete(key T) bool {
	i1, i2, _ := cf.getIndicesAndFingerprint(key, 0)
	if cf.buckets[i1].delete(key) || cf.buckets[i2].delete(key) {
		cf.count--
		return true
	}
	return false
}

// Decrement decrements the count of a key in the Counting Cuckoo Filter and returns the new count
func (cf *CountingCuckooFilter[T]) Decrement(key T) (bool, int) {
	i1, i2, _ := cf.getIndicesAndFingerprint(key, 0)
	if count := cf.buckets[i1].decrement(key); count >= 0 {
		cf.count--
		return true, int(count)
	}
	if count := cf.buckets[i2].decrement(key); count >= 0 {
		cf.count--
		return true, int(count)
	}
	return false, -1
}

// getIndicesAndFingerprint returns the two bucket indices and fingerprint for a key
func (cf *CountingCuckooFilter[T]) getIndicesAndFingerprint(key T, seed uint32) (uint32, uint32, uint16) {
	hash, _ := cf.hashFunction(key, seed)
	fp := cf.fingerprint(hash)
	i1 := hash % uint64(cf.numBuckets)
	i2 := cf.getAlternateIndex(fp, uint32(i1), seed)
	return uint32(i1), i2, fp
}

// getAlternateIndex returns the alternate bucket index for a fingerprint and index
func (cf *CountingCuckooFilter[T]) getAlternateIndex(fp uint16, i uint32, seed uint32) uint32 {
	hash := uint64(fp)
	return (i ^ uint32(hash)) % cf.numBuckets
}

// increment increments the count of a key in a specific bucket
func (cf *CountingCuckooFilter[T]) increment(key T, i uint32) bool {
	if cf.buckets[i].increment(key) {
		cf.count++
		return true
	}
	return false
}

// insert inserts a key into a specific bucket
func (cf *CountingCuckooFilter[T]) insert(key T, i uint32) bool {
	if cf.buckets[i].insert(key) {
		cf.count++
		return true
	}
	return false
}

// decrement decrements the count of a key in a specific bucket and returns the new count
func (cf *CountingCuckooFilter[T]) decrement(key T, i uint32) int {
	if count := cf.buckets[i].decrement(key); count >= 0 {
		cf.count--
		return count
	}
	return -1
}

// swapFingerprint swaps a fingerprint in a bucket and returns the swapped fingerprint
func (b *Bucket[T]) swapFingerprint(fp uint16) uint16 {
	for i := range b.keys {
		if b.counts[i] == 0 {
			b.keys[i], _ = any(fp).(T)
			b.counts[i] = 1
			return fp
		}
	}
	return fp
}

// getCount returns the count of a key in the bucket
func (b *Bucket[T]) getCount(key T) uint8 {
	for i := range b.keys {
		if b.counts[i] > 0 && common.EqualKeys(b.keys[i], key) {
			return b.counts[i]
		}
	}
	return 0
}

// increment increments the count of a key in the bucket
func (b *Bucket[T]) increment(key T) bool {
	for i := range b.keys {
		if b.counts[i] > 0 && common.EqualKeys(b.keys[i], key) {
			if b.counts[i] < maxCount {
				b.counts[i]++
				return true
			}
			return false
		}
	}
	return false
}

// delete removes a key from the bucket
func (b *Bucket[T]) delete(key T) bool {
	for i := range b.keys {
		if b.counts[i] > 0 && common.EqualKeys(b.keys[i], key) {
			b.counts[i] = 0
			b.keys[i] = b.empty
			return true
		}
	}
	return false
}

// decrement decrements the count of a key in the bucket and returns the new count
func (b *Bucket[T]) decrement(key T) int {
	for i := range b.keys {
		if b.counts[i] > 0 && common.EqualKeys(b.keys[i], key) {
			b.counts[i]--
			if b.counts[i] == 0 {
				b.keys[i] = b.empty
			}
			return int(b.counts[i])
		}
	}
	return -1
}

// insert inserts a key into the bucket
func (b *Bucket[T]) insert(key T) bool {
	for i := range b.keys {
		if b.counts[i] == 0 {
			b.keys[i] = key
			b.counts[i] = 1
			return true
		}
	}
	return false
}

// fingerprint returns the fingerprint of a hash
func (cf *CountingCuckooFilter[T]) fingerprint(hash uint64) uint16 {
	return uint16(hash & 0xFFFF)
}
