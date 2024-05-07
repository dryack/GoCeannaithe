package cuckoo

import (
	"encoding/binary"
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
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
/*func (cf *CuckooFilter[T]) Insert(key T) bool {
	i1, i2, fp := cf.getIndicesAndFingerprint(key, 0)

	if cf.insert(key, i1) || cf.insert(key, i2) {
		fmt.Println(key) // DEBUG
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
}*/

func (cf *CuckooFilter[T]) Insert(key T) bool {
	i1, i2, fp := cf.getIndicesAndFingerprint(key, 0)
	fpKey := any(fp).(T)

	if cf.insert(fpKey, i1) || cf.insert(fpKey, i2) {
		return true
	}

	i := i1
	for k := 0; k < maxKicks; k++ {
		oldKey := cf.buckets[i].swapFingerprint(fpKey)
		i = cf.getAlternateIndex(uint32(any(oldKey).(uint32)), i, 0)
		if cf.insert(oldKey, i) {
			return true
		}
		fpKey = oldKey
	}

	return false
}

// Lookup checks if a key exists in the Cuckoo Filter
func (cf *CuckooFilter[T]) Lookup(key T) bool {
	i1, i2, _ := cf.getIndicesAndFingerprint(key, 0)
	return cf.buckets[i1].contains(key) || cf.buckets[i2].contains(key)
}

// Delete removes a key from the Cuckoo Filter
func (cf *CuckooFilter[T]) Delete(key T) bool {
	i1, i2, _ := cf.getIndicesAndFingerprint(key, 0)
	if cf.buckets[i1].delete(key) || cf.buckets[i2].delete(key) {
		cf.count--
		return true
	}
	return false
}

// getIndicesAndFingerprint returns the two bucket indices and fingerprint for a key
func (cf *CuckooFilter[T]) getIndicesAndFingerprint(key T, seed uint32) (uint32, uint32, uint32) {
	keyBytes, _ := common.NumToBytes[T](key)
	// hash, _ := cf.hashFunction(key, seed)
	hash, _ := common.HashKeyMurmur3(keyBytes, seed)
	hashByes, _ := common.NumToBytes(hash)
	fp := cf.fingerprint(hashByes)
	i1 := hash % uint64(cf.numBuckets)
	fmt.Println("fp:", fp) // DEBUG
	fmt.Println("i1:", i1) // DEBUG
	i2 := cf.getAlternateIndex(fp, uint32(i1), seed)
	fmt.Println("i2:", i2) // DEBUG
	return uint32(i1), i2, fp
}

// getAlternateIndex returns the alternate bucket index for a fingerprint and index
func (cf *CuckooFilter[T]) getAlternateIndex(fp uint32, i uint32, seed uint32) uint32 {
	// hash, _ := cf.hashFunction(uint64(fp), seed)
	// hash, _ := common.HashKeyMurmur3[uint32](fp, seed)
	hash := uint64(fp)
	return i ^ uint32(hash)%cf.numBuckets
}

// insert inserts a key into a specific bucket
func (cf *CuckooFilter[T]) insert(key T, i uint32) bool {
	if cf.buckets[i].insert(key) {
		cf.count++
		return true
	}
	return false
}

// swapFingerprint swaps a fingerprint in a bucket and returns the swapped fingerprint
/*func (b *Bucket[T]) swapFingerprint(fp uint32) uint32 {
	// fmt.Println("swapFingerprint()")
	for i := range b.keys {
		/*if !b.occupied[i] {
			b.occupied[i] = true
			b.keys[i], _ = any(fp).(T)
			return fp
		}//*\/
		if b.occupied&(1<<i) == 0 {
			b.occupied |= 1 << i
			b.keys[i], _ = any(fp).(T)
			fmt.Println("occupied")
			fmt.Println(b.keys[i]) // DEBUG

			return fp
		}
		return fp
	}
	return fp
}*/

func (b *Bucket[T]) swapFingerprint(fp T) T {
	for i := range b.keys {
		if b.occupied&(1<<i) != 0 {
			oldKey := b.keys[i]
			b.keys[i] = fp
			return oldKey
		}
	}
	return fp
}

// contains checks if a bucket contains a key
func (b *Bucket[T]) contains(key T) bool {
	for i := range b.keys {
		/*if b.occupied[i] && common.EqualKeys(b.keys[i], key) {
			return true
		}*/
		if b.occupied&(1<<i) == 1 && common.EqualKeys(b.keys[i], key) {
			return true
		}
	}
	return false
}

// delete removes a key from a bucket
func (b *Bucket[T]) delete(key T) bool {
	for i := range b.keys {
		/*if b.occupied[i] && common.EqualKeys(b.keys[i], key) {
			b.occupied[i] = false
			b.keys[i] = b.empty
			return true
		}*/
		if b.occupied&(1<<i) == 1 && common.EqualKeys(b.keys[i], key) {
			b.occupied &^= 1 << i // unset the bit in the location int(i) of the byte b.occupied
			b.keys[i] = b.empty
			return true
		}
	}
	return false
}

// insert inserts a key into a bucket
func (b *Bucket[T]) insert(key T) bool {
	for i := range b.keys {
		/*if !b.occupied[i] {
			b.occupied[i] = true
			b.keys[i] = key
			return true
		}*/
		if b.occupied&(1<<i) == 0 {
			b.occupied |= 1 << i
			b.keys[i] = key
			return true
		}
	}
	return false
}

// fingerprint returns the fingerprint of a hash
func (cf *CuckooFilter[T]) fingerprint(hash []byte) uint32 {
	// fp := uint32(hash & 0xFFFF)
	fp := binary.BigEndian.Uint32(hash[:4]) & 0xFFFF
	fmt.Println(fp)
	fmt.Println(hash) // DEBUG
	return fp
}
