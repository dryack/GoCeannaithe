package cuckoo

import (
	"github.com/dryack/GoCeannaithe/pkg/common"
	"math/rand"
	"reflect"
	"unsafe"
)

const (
	maxKicks = 500
)

// Fingerprint is a type that determines the size of Bucket.fingerprints
type Fingerprint uint16

// Bucket represents a bucket in the Cuckoo Filter
type Bucket struct {
	fingerprints [4]Fingerprint

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
		buckets[i] = &Bucket{fingerprints: [4]Fingerprint{}, occupied: 0}
	}
	cf := &CuckooFilter[T]{
		buckets:    buckets,
		numBuckets: numBuckets,
		count:      0,
		capacity:   numBuckets * 4,
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
	hash, _ := cf.hashFunction(key, 0)
	fp := cf.fingerprint(hash)
	i1, i2 := cf.getIndicesAndFingerprint(hash, fp)

	if cf.insert(i1, fp) || cf.insert(i2, fp) {
		return true
	}

	i := i1
	for n := 0; n < maxKicks; n++ {
		f := cf.buckets[i].swapFingerprint(fp)
		fp = f
		i = cf.getAlternateIndex(fp, i, 0)

		if cf.insert(i, fp) {
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
	i2 := i1 ^ cf.hashFingerprint(fp)
	return i1, i2
}

// getAlternateIndex returns the alternate bucket index for a fingerprint and index
func (cf *CuckooFilter[T]) getAlternateIndex(fp Fingerprint, i uint32, seed uint32) uint32 {
	// hash, _ := cf.hashFunction(fp, seed)
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
	i := rand.Intn(len(b.fingerprints))
	for b.occupied&(1<<i) == 0 {
		i = (i + 1) % len(b.fingerprints)
	}
	oldFp := b.fingerprints[i]
	b.fingerprints[i] = fp
	return oldFp
}

// contains checks if a bucket contains a key using the fingerprint
func (b *Bucket) contains(fp Fingerprint) bool {
	for i := range b.fingerprints {
		if b.occupied&(1<<i) == 1 && b.fingerprints[i] == fp {
			return true
		}
	}
	return false
}

// insert inserts a key into a bucket using the fingerprint
func (b *Bucket) insert(fp Fingerprint) bool {
	for i := range b.fingerprints {
		if b.occupied&(1<<i) == 0 {
			b.occupied |= 1 << i
			b.fingerprints[i] = fp
			return true
		}
	}
	return false
}

// delete removes a key from a bucket using the fingerprint
func (b *Bucket) delete(fp Fingerprint) bool {
	for i := range b.fingerprints {
		if b.occupied&(1<<i) == 1 && b.fingerprints[i] == fp {
			b.occupied &^= 1 << i
			b.fingerprints[i] = 0
			return true
		}
	}
	return false
}

// fingerprint returns the fingerprint of a hash
func (cf *CuckooFilter[T]) fingerprint(hash uint64) Fingerprint {
	// fmt.Println(Fingerprint(hash >> (64 - 8*unsafe.Sizeof(Fingerprint(0))))) // DEBUG
	return Fingerprint(hash >> (64 - 8*unsafe.Sizeof(Fingerprint(0))))
}

func (cf *CuckooFilter[T]) hashFingerprint(fp Fingerprint) uint32 {
	hash, _ := cf.hashFunction(fp, 1234)
	return uint32(hash % uint64(cf.numBuckets))
}

// ApproximateSize returns the approximate size of the Cuckoo Filter in bytes
func (cf *CuckooFilter[T]) ApproximateSize() int64 {
	const (
		numFingerprints = 4 // Number of fingerprints per bucket
		occupiedSize    = 8 // occupied field takes 8 bytes (int64)
		pointerSize     = 8 // pointer size on 64-bit systems
	)

	// Get the type of the Bucket struct
	bucketType := reflect.TypeOf(cf.buckets[0]).Elem()

	// Get the type of the fingerprints field in the Bucket struct
	fingerprintsField, _ := bucketType.FieldByName("fingerprints")

	// Get the type of the fingerprint element
	fingerprintType := fingerprintsField.Type.Elem()

	// Get the size of the fingerprint type
	fingerprintSize := fingerprintType.Size()

	bucketSize := numFingerprints * fingerprintSize

	numBuckets := len(cf.buckets)
	bucketStructSize := bucketSize + occupiedSize
	totalBucketSize := int64(numBuckets) * (int64(bucketStructSize) + int64(pointerSize))

	filterMetadataSize := int64(4 * 3) // numBuckets, count, and hashEnum fields

	approximateSize := totalBucketSize + filterMetadataSize
	return approximateSize
}

func (cf *CuckooFilter[T]) GetLoadFactor() float64 {
	return float64(cf.count) / float64(cf.capacity)
}
