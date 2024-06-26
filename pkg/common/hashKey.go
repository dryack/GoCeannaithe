package common

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"github.com/dchest/siphash"
	"github.com/twmb/murmur3"
	"github.com/zeebo/xxh3"
)

const (
	UnknownHash = uint8(0)
	Murmur3     = uint8(1)
	Sha256      = uint8(2)
	Sha512      = uint8(3)
	SipHash     = uint8(4)
	XXhash      = uint8(5)
)

// HashKeyMurmur3 uses NumToBytes to convert a numeric type to bytes and computes the hash value using Murmur3.
func HashKeyMurmur3[T Hashable](key T, seed uint32) (uint64, error) {
	keyBytes, err := NumToBytes[T](key) // Directly using NumToBytes
	if err != nil {
		return 0, err // Properly handle errors from NumToBytes
	}

	h1, h2 := murmur3.SeedSum128(uint64(seed), uint64(seed), keyBytes)
	return h1 ^ h2, nil
}

// HashKeySha256 uses NumToBytes to convert a numeric type to bytes and computes the hash value using SHA-256
func HashKeySha256[T Hashable](key T, seed uint32) (uint64, error) {
	keyBytes, err := NumToBytes[T](key)
	if err != nil {
		return 0, err
	}

	// Prepare buffer and write seed and key to it
	buffer := make([]byte, 4+len(keyBytes))
	binary.BigEndian.PutUint32(buffer[:4], seed)
	copy(buffer[4:], keyBytes)

	sum := sha256.Sum256(buffer)

	// Convert the first and second 8 bytes of hash output to uint64
	h1 := binary.BigEndian.Uint64(sum[:8])
	h2 := binary.BigEndian.Uint64(sum[8:16])

	return h1 ^ h2, nil
}

// HashKeySha512 uses NumToBytes to convert a numeric type to bytes and computes the hash value using SHA-512
func HashKeySha512[T Hashable](key T, seed uint32) (uint64, error) {
	keyBytes, err := NumToBytes[T](key)
	if err != nil {
		return 0, err
	}

	// Prepare buffer and write seed and key to it
	buffer := make([]byte, 4+len(keyBytes))
	binary.BigEndian.PutUint32(buffer[:4], seed)
	copy(buffer[4:], keyBytes)

	sum := sha512.Sum512(buffer)

	// Convert the first 8 bytes of hash output to uint64
	value := binary.BigEndian.Uint64(sum[:8])

	return value, nil
}

// HashKeySipHash uses NumToBytes to convert a numeric type to bytes and computes the hash value using SipHash
func HashKeySipHash[T Hashable](key T, seed uint32) (uint64, error) {
	keyBytes, err := NumToBytes[T](key)
	if err != nil {
		return 0, err
	}

	h1, h2 := siphash.Hash128(uint64(seed), uint64(seed), keyBytes)

	return h1 ^ h2, nil
}

// HashKeyXXhash uses NumToBytes to convert a numeric type to bytes and computes the hash value using XX Hash
func HashKeyXXhash[T Hashable](key T, seed uint32) (uint64, error) {
	keyBytes, err := NumToBytes[T](key)
	if err != nil {
		return 0, err
	}

	h := xxh3.Hash128Seed(keyBytes, uint64(seed))

	return h.Hi ^ h.Lo, nil
}
