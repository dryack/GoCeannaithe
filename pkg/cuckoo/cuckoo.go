package cuckoo

import (
	"github.com/dryack/GoCeannaithe/pkg/common"
	"math/rand"
)

const (
	maxKicks = 500
)

type Fingerprint uint16

// Bucket represents a bucket in the Cuckoo Filter
type Bucket struct {
	data uint64

	/*
		TODO: the raw array of bools currently provides the best performance on
			insertions, while using bit manipulation (especially int64) will provide the
			best performance when doing the other commands. If this remains true after
			the next bout of testing, do we want to permit the end-user to choose which
			approach they want - and if so, how?
	*/

}

const (
	fingerprintBits = 13
	counterBits     = 2
	fingerprints    = 4
	occupiedMask    = 0xF000000000000000
	fingerprintMask = 0x1FFF
	counterMask     = 0x3
)

// CuckooFilter represents a Cuckoo Filter
type CuckooFilter[T common.Hashable] struct {
	buckets      []*Bucket
	numBuckets   uint32
	count        uint32
	capacity     uint32
	hashFunction func(any, uint32) (uint64, error)
	hashEnum     uint8
}

// NewCuckooFilter creates a new Cuckoo Filter with the specified number of buckets
func NewCuckooFilter[T common.Hashable](numBuckets uint32) *CuckooFilter[T] {
	buckets := make([]*Bucket, numBuckets)
	for i := range buckets {
		buckets[i] = &Bucket{data: 0}
	}
	cf := &CuckooFilter[T]{
		buckets:    buckets,
		numBuckets: numBuckets,
		count:      0,
		capacity:   numBuckets * 4,
	}
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
	hash, _ := cf.hashFunction(key, 0)
	fp := cf.fingerprint(hash)
	i1, i2 := cf.getIndicesAndFingerprint(hash, fp)

	if cf.buckets[i1].insert(fp) || cf.buckets[i2].insert(fp) {
		cf.count++
		return true
	}

	i := i1
	for n := 0; n < maxKicks; n++ {
		f := cf.buckets[i].swapFingerprint(fp)
		fp = f
		i = cf.getAlternateIndex(fp, i, 0)

		if cf.buckets[i].insert(fp) {
			cf.count++
			return true
		}
	}

	return false
}

// Lookup checks if a key exists in the Cuckoo Filter
func (cf *CuckooFilter[T]) Lookup(key T) bool {
	hash, _ := cf.hashFunction(key, 0)
	fp := cf.fingerprint(hash)
	i1, i2 := cf.getIndicesAndFingerprint(hash, fp)
	return cf.buckets[i1].contains(fp) || cf.buckets[i2].contains(fp)
}

// Delete removes a key from the Cuckoo Filter
func (cf *CuckooFilter[T]) Delete(key T) bool {
	hash, _ := cf.hashFunction(key, 0)
	fp := cf.fingerprint(hash)
	i1, i2 := cf.getIndicesAndFingerprint(hash, fp)
	if cf.buckets[i1].delete(fp) || cf.buckets[i2].delete(fp) {
		cf.count--
		return true
	}
	return false
}

// getIndicesAndFingerprint returns the two bucket indices and fingerprint for a key
func (cf *CuckooFilter[T]) getIndicesAndFingerprint(hash uint64, fp Fingerprint) (uint32, uint32) {
	i1 := uint32(hash % uint64(cf.numBuckets))
	hash2, _ := cf.hashFunction(fp, 5678)
	i2 := uint32(hash2 % uint64(cf.numBuckets))
	return i1, i2
}

// getAlternateIndex returns the alternate bucket index for a fingerprint and index
func (cf *CuckooFilter[T]) getAlternateIndex(fp Fingerprint, i uint32, seed uint32) uint32 {
	hash, _ := cf.hashFunction(fp, seed+5678)
	return uint32(hash % uint64(cf.numBuckets))
}

// insert inserts a key into a specific bucket using the fingerprint
func (cf *CuckooFilter[T]) insert(i uint32, fp Fingerprint) bool {
	if cf.buckets[i].insert(fp) {
		// fmt.Println(cf.buckets[i].fingerprints) // DEBUG
		cf.count++
		// fmt.Println(cf.count) // DEBUG
		return true
	}
	return false
}

// swapFingerprint swaps a fingerprint in a bucket and returns the swapped fingerprint
func (b *Bucket) swapFingerprint(fp Fingerprint) Fingerprint {
	i := rand.Intn(fingerprints)
	for b.data&(occupiedMask>>(i*15)) == 0 {
		i = (i + 1) % fingerprints
	}
	oldFp := Fingerprint((b.data >> (i * 15)) & fingerprintMask)
	b.data &^= uint64(fingerprintMask << (i * 15))
	b.data |= uint64(fp) << (i * 15)
	return oldFp
}

// contains checks if a bucket contains a key using the fingerprint
func (b *Bucket) contains(fp Fingerprint) bool {
	for i := 0; i < fingerprints; i++ {
		if b.data&(occupiedMask>>(i*15)) != 0 {
			storedFp := Fingerprint((b.data >> (i * 15)) & fingerprintMask)
			if storedFp == fp {
				return true
			}
		}
	}
	return false
}

// insert inserts a key into a bucket using the fingerprint
func (b *Bucket) insert(fp Fingerprint) bool {
	for i := 0; i < fingerprints; i++ {
		if b.data&(occupiedMask>>(i*15)) != 0 {
			storedFp := Fingerprint((b.data >> (i * 15)) & fingerprintMask)
			if storedFp == fp {
				count := (b.data >> ((i * 15) + fingerprintBits)) & counterMask
				if count < counterMask {
					b.data += 1 << ((i * 15) + fingerprintBits)
					return true
				}
			}
		} else {
			b.data |= occupiedMask >> (i * 15)
			b.data &^= uint64(fingerprintMask << (i * 15))
			b.data |= uint64(fp) << (i * 15)
			return true
		}
	}
	return false
}

// delete removes a key from a bucket using the fingerprint
func (b *Bucket) delete(fp Fingerprint) bool {
	for i := 0; i < fingerprints; i++ {
		if b.data&(occupiedMask>>(i*15)) != 0 {
			storedFp := Fingerprint((b.data >> (i * 15)) & fingerprintMask)
			if storedFp == fp {
				count := (b.data >> ((i * 15) + fingerprintBits)) & counterMask
				if count > 0 {
					b.data -= 1 << ((i * 15) + fingerprintBits)
				} else {
					b.data &^= occupiedMask >> (i * 15)
					b.data &^= uint64(fingerprintMask << (i * 15))
				}
				return true
			}
		}
	}
	return false
}

// fingerprint returns the fingerprint of a hash
func (cf *CuckooFilter[T]) fingerprint(hash uint64) Fingerprint {
	return Fingerprint(hash & fingerprintMask)
}

func (cf *CuckooFilter[T]) hashFingerprint(fp Fingerprint) uint32 {
	hash, _ := cf.hashFunction(fp, 1234)
	return uint32(hash % uint64(cf.numBuckets))
}

// FIXME:  currently broken
// ApproximateSize returns the approximate size of the Cuckoo Filter in bytes
func (cf *CuckooFilter[T]) ApproximateSize() int64 {
	const (
		numFingerprints = 4  // Number of fingerprints per bucket
		bucketSize      = 64 // Size of the bucket in bits (4 * (13 + 2) + 4)
		pointerSize     = 8  // Pointer size in bytes (64-bit systems)
	)

	numBuckets := len(cf.buckets)
	totalBucketSize := int64(numBuckets) * (bucketSize/8 + pointerSize)

	filterMetadataSize := int64(4 * 3) // numBuckets, count, and hashEnum fields

	approximateSize := totalBucketSize + filterMetadataSize
	return approximateSize
}

func (cf *CuckooFilter[T]) GetLoadFactor() float64 {
	return float64(cf.count) / float64(cf.capacity)
}
