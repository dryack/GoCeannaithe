package common

// Hashable is an interface that represents types that can be hashed.
type Hashable interface {
	~string | ~[]byte | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}
