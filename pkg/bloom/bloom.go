package bloom

import (
	"errors"
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"math"
	"math/bits"
)

// Storage defines the interface for bit storage operations
type Storage[T common.Hashable] interface {
	SetBit(key T) error
	CheckBit(key T) bool
}

// BitPackingStorage uses a slice of uint64 for efficient bit storage
type BitPackingStorage[T common.Hashable] struct {
	bits       []uint64
	seeds      []uint32 // TODO: We may want to just make these uint64, and avoid casting them when hashing
	bitsLength uint64
}

// ConventionalStorage uses a slice of bool
type ConventionalStorage[T common.Hashable] struct {
	bits        []bool
	seeds       []uint32 // TODO: We may want to just make these uint64, and avoid casting them when hashing
	sliceLength uint64
}

// NewBitPackingStorage creates a new BitPackingStorage with the given number of bits
//
// Size here indicates the number of bits, and not the number of keys we wish to store.
func NewBitPackingStorage[T common.Hashable](size uint64, seeds []uint32) *BitPackingStorage[T] {
	roundedSize := roundUpToNextPowerOfTwo(size)
	numUint64s := (roundedSize + 63) / 64 // Calculate number of uint64s needed
	return &BitPackingStorage[T]{bits: make([]uint64, numUint64s), seeds: seeds, bitsLength: numUint64s}
}

// NewConventionalStorage creates a new ConventionalStorage with the specified size
//
// Size here indicates the number of bits - or here, being ConventionalStorage) - the number of cells in the slice, and
// not the number of keys we wish to store.
func NewConventionalStorage[T common.Hashable](size uint64, seeds []uint32) *ConventionalStorage[T] {
	return &ConventionalStorage[T]{bits: make([]bool, size), seeds: seeds, sliceLength: size}
}

// calculateBitIndex calculates the bit index for a given key
func (b *BitPackingStorage[T]) calculateBitIndex(key T, seed uint32) (uint64, error) {
	index, err := common.HashKey[T](key, seed)
	if err != nil {
		return 0, err
	}
	// return index % uint64(len(b.bits)*64), nil
	return index % (b.bitsLength * 64), nil
}

// SetBit sets the bit at the calculated index for a given key by iterating
// through each seed in the BitPackingStorage. It uses the calculateBitIndex
// method to determine the index and then sets the bit at that index using
// bitwise operations. If there is an error during the index calculation,
// the error is returned.
func (b *BitPackingStorage[T]) SetBit(key T) error {
	for _, seed := range b.seeds {
		index, err := b.calculateBitIndex(key, seed)
		if err != nil {
			return err
		}
		b.bits[index/64] |= 1 << (index % 64)
	}
	return nil
}

// CheckBit checks if all bits corresponding to the given key are set to true.
// It iterates through the seeds and calculates the bit index using the calculateBitIndex method.
// If there is an error or the bit at the calculated index is not set, it returns false.
// Otherwise, it returns true.
func (b *BitPackingStorage[T]) CheckBit(key T) bool {
	for _, seed := range b.seeds {
		index, _ := b.calculateBitIndex(key, seed)
		if (b.bits[index/64] & (1 << (index % 64))) == 0 {
			return false // the bit for this hash/seed is not set
		}
	}
	return true
}

// calculateBitIndex calculates the bit index for a given key in ConventionalStorage
func (c *ConventionalStorage[T]) calculateBitIndex(key T, seed uint32) (uint64, error) {
	index, err := common.HashKey[T](key, seed)
	if err != nil {
		return 0, err
	}
	return index % c.sliceLength, nil
}

// SetBit sets the bit at the calculated index for a given key by iterating
// through each seed in the ConventionalStorage. It uses the calculateBitIndex
// method to determine the index and then sets the bit at that index using
// bitwise operations. If there is an error during the index calculation,
// the error is returned.
func (c *ConventionalStorage[T]) SetBit(key T) error {
	for _, seed := range c.seeds {
		index, err := c.calculateBitIndex(key, seed)
		if err != nil {
			fmt.Println(err) // DEBUG
			return err
		}
		// fmt.Printf("Setting bit at index %d for key %v using seed %d\n", index, key, seed) // DEBUG
		c.bits[index] = true
	}
	return nil
}

// CheckBit checks if all bits corresponding to the given key are set to true.
// It iterates through the seeds and calculates the bit index using the calculateBitIndex method.
// If there is an error or the bit at the calculated index is not set, it returns false.
// Otherwise, it returns true.
func (c *ConventionalStorage[T]) CheckBit(key T) bool {
	for _, seed := range c.seeds {
		index, err := c.calculateBitIndex(key, seed)
		if err != nil || !c.bits[index] {
			// fmt.Printf("Bit at index %d for key %v using seed %d is not set\n", index, key, seed) // DEBUG
			return false
		}
	}
	return true
}

// BloomFilter holds the bit storage and hash functions
type BloomFilter[T common.Hashable] struct {
	Storage          Storage[T]
	numHashFunctions int
	seeds            []uint32
}

// NewBloomFilter creates a new BloomFilter, initially with no storage
func NewBloomFilter[T common.Hashable]() *BloomFilter[T] {
	return &BloomFilter[T]{}
}

// WithStorage sets the storage mechanism for the BloomFilter
func (bf *BloomFilter[T]) WithStorage(storage Storage[T]) (*BloomFilter[T], error) {
	// redundant currently, but i can imagine each case having different requirements in the future
	switch s := storage.(type) {
	case *BitPackingStorage[T]:
		s.seeds = bf.seeds
	case *ConventionalStorage[T]:
		s.seeds = bf.seeds
	default:
		err := errors.New("unsupported storage type")
		return nil, err
	}
	bf.Storage = storage
	return bf, nil
}

// TODO:  I'd really like WithAutoConfigure to return the details of all 4 parameters to the user

// WithAutoConfigure may be used in lieu of WithHashFunctions and WithStorage. It does this by calculating the best parameters for the
// Bloom Filter, based upon the formulas:
//
// “m = -n * ln(p) / (ln(2)^2)“, where m is the number of bits, n is the required number of elements to be stored, and p
// is the requested error-rate, and
//
// “k = (m / n) * ln(2)“, where k is the number of key hashes to use
//
// It picks the most memory efficient Storage option (which will almost always be BitPackingStorage unless an
// tiny number of elements are expected to be stored in the Bloom Filter
func (bf *BloomFilter[T]) WithAutoConfigure(elements uint64, requestedErrorRate float64) (*BloomFilter[T], error) {
	m := int(math.Ceil(-float64(elements) * math.Log(requestedErrorRate) / (math.Ln2 * math.Ln2)))
	k := int(math.Ceil((float64(m) / float64(elements)) * math.Ln2))

	// Initialize the seeds array for hash functions
	seeds := make([]uint32, k)
	for i := range seeds {
		seeds[i] = uint32(i + 1) // Simple sequence of seeds, ensure non-zero and simple variability
	}

	var storage Storage[T]
	if m < 64 { // TODO: in theory we could check the size of bool with unsafe.Sizeof() just in case it's ever larger than a byte..
		storage = NewConventionalStorage[T](uint64(m), seeds)
	} else {
		storage = NewBitPackingStorage[T](uint64(m), seeds)
	}

	// Set the storage and hash functions
	bf.Storage = storage
	bf.numHashFunctions = k
	bf.seeds = seeds

	return bf, nil
}

// TODO: WithHashFunctions probably should probably return different functions, thereby allowing the user to select which hashing function they want to use, and how the seeds are defined

// WithHashFunctions sets the number of hash functions to use and initializes the seeds
func (bf *BloomFilter[T]) WithHashFunctions(h int) *BloomFilter[T] {
	bf.numHashFunctions = h
	bf.seeds = make([]uint32, h)
	for i := range bf.seeds {
		bf.seeds[i] = uint32(i) // TODO: break this out to allow different methods of creating seed values
	}

	switch storage := bf.Storage.(type) {
	case *BitPackingStorage[T]:
		storage.seeds = bf.seeds
	case *ConventionalStorage[T]:
		storage.seeds = bf.seeds
	}
	return bf
}

// roundUpToNextPowerOfTwo finds the next power of two value for a given number
func roundUpToNextPowerOfTwo(x uint64) uint64 {
	if x < 1 {
		return 1
	}
	return 1 << (bits.Len(uint(x - 1))) // Use bits.Len to find the next power of two
}
