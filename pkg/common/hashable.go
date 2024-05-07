package common

import "bytes"

// Hashable is an interface that represents types that can be hashed.
type Hashable interface {
	~string | ~[]byte | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// qualKeys compares two keys for equality based on their underlying type
func EqualKeys[T Hashable](key1, key2 T) bool {
	switch k1 := any(key1).(type) {
	case string:
		k2, ok := any(key2).(string)
		return ok && k1 == k2
	case []byte:
		k2, ok := any(key2).([]byte)
		return ok && bytes.Equal(k1, k2)
	case int:
		k2, ok := any(key2).(int)
		return ok && k1 == k2
	case int8:
		k2, ok := any(key2).(int8)
		return ok && k1 == k2
	case int16:
		k2, ok := any(key2).(int16)
		return ok && k1 == k2
	case int32:
		k2, ok := any(key2).(int32)
		return ok && k1 == k2
	case int64:
		k2, ok := any(key2).(int64)
		return ok && k1 == k2
	case uint:
		k2, ok := any(key2).(uint)
		return ok && k1 == k2
	case uint8:
		k2, ok := any(key2).(uint8)
		return ok && k1 == k2
	case uint16:
		k2, ok := any(key2).(uint16)
		return ok && k1 == k2
	case uint32:
		k2, ok := any(key2).(uint32)
		return ok && k1 == k2
	case uint64:
		k2, ok := any(key2).(uint64)
		return ok && k1 == k2
	case float32:
		k2, ok := any(key2).(float32)
		return ok && k1 == k2
	case float64:
		k2, ok := any(key2).(float64)
		return ok && k1 == k2
	default:
		return false
	}
}
