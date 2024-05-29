package cuckoo2

import (
	"errors"
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"math/rand"
)

type Fingerprint uint16

type Bucket struct {
	data BitField
}

type CuckooFilter[T common.Hashable] struct {
	buckets      []*Bucket
	numBuckets   uint32
	hashPower    uint32 // 2**hashPower == numBuckets TODO: this may not be necessary if we're not going to dynamically size the filter
	count        uint32
	capacity     uint64
	victimCache  [16]Fingerprint // TODO: probably will be changed
	bm           *BitManipulator
	hashFunction func(any, uint32) (uint64, error)
	hashEnum     uint8
	maxKicks     uint16
}

func NewCuckooFilter[T common.Hashable]() *CuckooFilter[T] {
	return &CuckooFilter[T]{}
}

// WithHash returns a pointer to a CuckooFilter configured to use the specified hash function.
//
// The following hash functions are currently supported:
// common.Murmur3
//
// common.Sha256
//
// common.Sha512
//
// common.SipHash
//
// common.XXhash
func (cf *CuckooFilter[T]) WithHash(hashFunc uint8) (*CuckooFilter[T], error) {
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

		msg := fmt.Sprintf("Hash function not recognized: %v", hashFunc)
		return nil, errors.New(msg)
	}
	cf.hashEnum = hashFunc
	return cf, nil
}

// WithAutoConfigure returns a CuckooFilter that has been configured with
// reasonable defaults for its parameter numItems, the number of keys the user
// expects to store in the filter.
func (cf *CuckooFilter[T]) WithAutoConfigure(numItems uint32) (*CuckooFilter[T], float64, int64, uint32) {
	return &CuckooFilter[T]{}, 0, 0, 0
}

// Insert attempts to insert key into the filter.
func (cf *CuckooFilter[T]) Insert(key T) bool {
	hash, err := cf.hashFunction(key, 0)
	if err != nil {
		msg := fmt.Sprintf("Error hashing %v: %v, this is likely a bug", key, err)
		panic(msg)
	}

	fingerprint := cf.fingerprint(hash)
	idx1, idx2 := cf.getIndices(hash, fingerprint)

	if inserted := cf.buckets[idx1].insert(fingerprint, *cf.bm); inserted {
		cf.count++
		return true
	}

	if inserted := cf.buckets[idx2].insert(fingerprint, *cf.bm); inserted {
		cf.count++
		return true
	}

	// start kicking out fingerprints to find space for this insertion
	idx := idx1

	for n := 0; n < int(cf.maxKicks); n++ {
		if inserted := cf.buckets[idx1].insert(fingerprint, *cf.bm); inserted {
			cf.count++
			return true
		}

		if inserted := cf.buckets[idx2].insert(fingerprint, *cf.bm); inserted {
			cf.count++
			return true
		}

	}

	// swap the fingerprint with a randomly selected entry from the bucket
	if !cf.buckets[idx].insert(fingerprint, *cf.bm) {
		// Randomly select an entry from the bucket
		entryIndex := rand.Intn(int(cf.bm.FingerprintCount))
		entryFP := cf.bm.GetFingerprint(cf.buckets[idx].data, entryIndex)

		// Swap the fingerprints
		cf.bm.SetFingerprint(&cf.buckets[idx].data, entryIndex, uint64(fingerprint))
		fingerprint = Fingerprint(entryFP)

		// calculate the alternate bucket index using XOR
		hash, err := cf.hashFunction(uint16(fingerprint), 0)
		if err != nil {
			msg := fmt.Sprintf("Error hashing %v: %v, this is probably a bug", fingerprint, err)
			panic(msg)
		}
		idx = (idx ^ uint32(hash)) % cf.numBuckets

		if cf.buckets[idx].insert(fingerprint, *cf.bm) {
			cf.count++
			return true
		}
	} else {
		cf.count++
		return true
	}

	return false
}

// Delete attempts to delete key from the filter.
//
// It's essential to never attempt to delete a key that has never been inserted.
func (cf *CuckooFilter[T]) Delete(key T) bool {
	return false
}

// Lookup checks to see if key is found in the filter; by design false positives can happen during lookups.
func (cf *CuckooFilter[T]) Lookup(key T) bool {
	return false
}

func (cf *CuckooFilter[T]) fingerprint(hash uint64) Fingerprint {
	const (
		Shift33       = 33
		MixingFactor1 = 0xff51afd7ed558ccd
		MixingFactor2 = 0xc4ceb9fe1a85ec53
	)
	mixed := hash ^ (hash >> Shift33)
	mixed = (mixed * MixingFactor1) ^ ((mixed * MixingFactor2) >> Shift33)
	mixed = (mixed ^ (mixed >> Shift33)) * MixingFactor2
	mixed = mixed ^ (mixed >> Shift33)
	return Fingerprint(mixed & uint64(cf.bm.FingerprintMask))
}

func (cf *CuckooFilter[T]) getIndices(hash uint64, fp Fingerprint) (uint32, uint32) {
	i1 := uint32(hash % uint64(cf.numBuckets))
	hash2, err := cf.hashFunction(uint16(fp), 0)
	if err != nil {
		msg := fmt.Sprintf("error getting hash with fingerprint %d, in getIndices: %v, this is probably a bug", fp, err)
		panic(msg)
	}
	i2 := (i1 ^ uint32(hash2)) % uint32(cf.numBuckets)
	return i1, i2
}

// TODO: not sure we need this anymore
func (cf *CuckooFilter[T]) getAlternateIndex(fp Fingerprint, index uint32) uint32 {
	hash, err := cf.hashFunction(uint16(fp), 0)
	if err != nil {
		msg := fmt.Sprintf("Error hashing %v: %v", fp, err)
		panic(msg)
	}
	return uint32(hash % uint64(cf.numBuckets))
}

func (b *Bucket) insert(fp Fingerprint, bm BitManipulator) bool {
	for i := 0; uint32(i) < bm.FingerprintCount; i++ {
		if !bm.IsInUse(b.data, i) {
			bm.SetFingerprint(&b.data, i, uint64(fp))
			bm.SetInUse(&b.data, i, true)
			return true
		}
	}
	return false
}
