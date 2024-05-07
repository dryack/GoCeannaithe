package cuckoo

import (
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"math/rand"
)

const (
	maxKicks = 500
)

// Bucket represents a bucket in the Cuckoo Filter
type Bucket[T common.Hashable] struct {
	keys  [4]T
	empty T

	/*
		TODO: the raw array of bools currently provides the best performance on
			insertions, while using bit manipulation (especially int64) will provide the
			best performance when doing the other commands. If this remains true after
			the next bout of testing, do we want to permit the end-user to choose which
			approach they want - and if so, how?
	*/

	// occupied [4]bool
	occupied int64
}

// CuckooFilter represents a Cuckoo Filter
type CuckooFilter[T common.Hashable] struct {
	buckets      []*Bucket[T]
	numBuckets   uint32
	count        uint32
	hashFunction func(any, uint32) (uint64, error)
	hashEnum     uint8
}

// NewCuckooFilter creates a new Cuckoo Filter with the specified number of buckets
func NewCuckooFilter[T common.Hashable](numBuckets uint32) *CuckooFilter[T] {
	buckets := make([]*Bucket[T], numBuckets)
	for i := range buckets {
		buckets[i] = &Bucket[T]{empty: common.ZeroValue[T]()}
	}
	cf := &CuckooFilter[T]{
		buckets:    buckets,
		numBuckets: numBuckets,
		count:      0,
	}
	// cf.WithHashFunction(hashFunc) move this to cf.WithAutoConfigure when ready
	return cf
}

// WithHashFunction sets the hash function to be used by the Cuckoo Filter
func (cf *CuckooFilter[T]) WithHashFunction(hashFunc uint8) *CuckooFilter[T] {
	switch hashFunc {
	case common.Murmur3:
		cf.hashFunction = common.HashKeyMurmur3
	case common.Sha256:
		cf.hashFunction = common.HashKeySha256
	case common.Sha512:
		cf.hashFunction = common.HashKeySha512
	case common.SipHash:
		cf.hashFunction = common.HashKeySipHash
	case common.XXhash:
		cf.hashFunction = common.HashKeyXXhash
	default:
		cf.hashFunction = common.HashKeyMurmur3
	}
	cf.hashEnum = hashFunc
	return cf
}

// Insert inserts a key into the Cuckoo Filter
func (cf *CuckooFilter[T]) Insert(key T) bool {
	i1, i2, fp := cf.getIndicesAndFingerprint(key, 0)

	if cf.insert(key, i1, fp) || cf.insert(key, i2, fp) {
		return true
	}

	i := i1
	for k := 0; k < maxKicks; k++ {
		f := cf.buckets[i].swapFingerprint(fp, key)
		fp = f

		i = cf.getAlternateIndex(fp, i, 0)
		if cf.insert(key, i, fp) {
			return true
		}
	}
	// fmt.Println("no empty slot found") // DEBUG
	return false
}

// Lookup checks if a key exists in the Cuckoo Filter
func (cf *CuckooFilter[T]) Lookup(key T) bool {
	i1, i2, fp := cf.getIndicesAndFingerprint(key, 0)
	return cf.buckets[i1].contains(key, fp) || cf.buckets[i2].contains(key, fp)
}

// Delete removes a key from the Cuckoo Filter
func (cf *CuckooFilter[T]) Delete(key T) bool {
	i1, i2, fp := cf.getIndicesAndFingerprint(key, 0)
	if cf.buckets[i1].delete(key, fp) || cf.buckets[i2].delete(key, fp) {
		cf.count--
		return true
	}
	return false
}

// getIndicesAndFingerprint returns the two bucket indices and fingerprint for a key
func (cf *CuckooFilter[T]) getIndicesAndFingerprint(key T, seed uint32) (uint32, uint32, uint32) {
	hash, _ := cf.hashFunction(key, seed)
	fp := cf.fingerprint(hash)
	i1 := uint32(hash % uint64(cf.numBuckets))
	i2 := cf.getAlternateIndex(fp, i1, seed)
	return i1, i2, fp
}

// getAlternateIndex returns the alternate bucket index for a fingerprint and index
func (cf *CuckooFilter[T]) getAlternateIndex(fp uint32, i uint32, seed uint32) uint32 {
	hash, _ := cf.hashFunction(fp, seed)
	return i ^ uint32(hash)%cf.numBuckets
}

// insert inserts a key into a specific bucket using the fingerprint
func (cf *CuckooFilter[T]) insert(key T, i uint32, fp uint32) bool {
	if cf.buckets[i].insert(key, fp) {
		fmt.Println(cf.buckets[i].keys) // DEBUG
		cf.count++
		fmt.Println(cf.count) // DEBUG
		return true
	}
	return false
}

// swapFingerprint swaps a fingerprint in a bucket and returns the swapped fingerprint
func (b *Bucket[T]) swapFingerprint(fp uint32, key T) uint32 {
	i := rand.Intn(len(b.keys))
	oldKey := b.keys[i]
	oldFp := b.fingerprint(oldKey)
	b.keys[i] = key
	return oldFp
}

// contains checks if a bucket contains a key using the fingerprint
func (b *Bucket[T]) contains(key T, fp uint32) bool {
	for i := range b.keys {
		if b.occupied&(1<<i) == 1 && b.fingerprint(b.keys[i]) == fp && common.EqualKeys(b.keys[i], key) {
			return true
		}
	}
	return false
}

// delete removes a key from a bucket using the fingerprint
func (b *Bucket[T]) delete(key T, fp uint32) bool {
	for i := range b.keys {
		if b.occupied&(1<<i) == 1 && b.fingerprint(b.keys[i]) == fp && common.EqualKeys(b.keys[i], key) {
			b.occupied &^= 1 << i
			b.keys[i] = b.empty
			return true
		}
	}
	return false
}

// insert inserts a key into a bucket using the fingerprint
func (b *Bucket[T]) insert(key T, fp uint32) bool {
	for i := range b.keys {
		if b.occupied&(1<<i) == 0 {
			b.occupied |= 1 << i
			b.keys[i] = key
			return true
		}
	}
	return false
}

// fingerprint returns the fingerprint of a hash
func (cf *CuckooFilter[T]) fingerprint(hash uint64) uint32 {
	return uint32(hash & 0xFFFF)
}

// fingerprint returns the fingerprint of a key
func (b *Bucket[T]) fingerprint(key T) uint32 {
	// TODO: probably want to ensure we're using the same hash as the cuckoo filter's hashFunc
	hash, _ := common.HashKeyMurmur3(key, 0)
	return uint32(hash & 0xFFFF)
}
