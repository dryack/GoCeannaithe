package common

import "github.com/twmb/murmur3"

// TODO:  support additional hashing methods beyond Murmur3

// HashKey uses NumToBytes to convert a numeric type to bytes and computes the hash value using Murmur3.
func HashKey[T Hashable](key T, seed uint32) (uint64, error) {
	keyBytes, err := NumToBytes[T](key) // Directly using NumToBytes
	if err != nil {
		return 0, err // Properly handle errors from NumToBytes
	}

	h1, h2 := murmur3.SeedSum128(uint64(seed), uint64(seed), keyBytes)
	return h1 ^ h2, nil
}

/*func HashKey[T Hashable](key T, seed uint32) (uint64, error) {
	// fmt.Println("entering hash key") // DEBUG
	var keyBytes []byte

	switch k := any(key).(type) {
	case string:
		keyBytes = []byte(k)
	case []byte:
		keyBytes = k
	case int:
		keyBytes, _ = NumToBytes[int](k)
	case uint:
		keyBytes, _ = NumToBytes[uint](k)
	case int8:
		keyBytes, _ = NumToBytes[int8](k)
	case int16:
		keyBytes, _ = NumToBytes[int16](k)
	case int32:
		keyBytes, _ = NumToBytes[int32](k)
	case uint8:
		keyBytes, _ = NumToBytes[uint8](k)
	case uint16:
		keyBytes, _ = NumToBytes[uint16](k)
	case uint32:
		keyBytes, _ = NumToBytes[uint32](k)
	case int64:
		keyBytes, _ = NumToBytes[int64](k)
	case uint64:
		keyBytes, _ = NumToBytes[uint64](k)
	case float32:
		keyBytes, _ = NumToBytes[float32](k)
	case float64:
		keyBytes, _ = NumToBytes[float64](k)
	default:
		return 0, errors.New("unsupported type")
	}

	h1, h2 := murmur3.SeedSum128(uint64(seed), uint64(seed), keyBytes)

	return h1 ^ h2, nil
}
*/
