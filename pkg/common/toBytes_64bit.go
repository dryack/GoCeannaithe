// file: toBytes_64bit.go
//go:build amd64 || arm64

package common

import (
	"encoding/binary"
	"errors"
	"math"
)

func NumToBytes(num any) ([]byte, error) {
	switch k := num.(type) {
	case int:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(k))
		return buf, nil
	case uint:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(k))
		return buf, nil
	case int8:
		return []byte{byte(k)}, nil
	case uint8:
		return []byte{k}, nil
	case int16:
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(k))
		return buf, nil
	case uint16:
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, k)
		return buf, nil
	case int32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(k))
		return buf, nil
	case uint32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, k)
		return buf, nil
	case int64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(k))
		return buf, nil
	case uint64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, k)
		return buf, nil
	case float32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, math.Float32bits(k))
		return buf, nil
	case float64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(k))
		return buf, nil
	case string:
		return []byte(k), nil
	case []byte:
		return k, nil
	default:
		return nil, errors.New("unsupported type for binary conversion")
	}
}

/*
Changes from the commented out version of NumToBytes to the active one gave the following results:

1. Reduction in Time per Operation (ns/op):
BitPackingStorage:
SetBit improved from 689.5 ns/op to 169.1 ns/op.
CheckBit improved from 245.8 ns/op to 72.51 ns/op.
ConventionalStorage:
SetBit improved from 740.8 ns/op to 195.6 ns/op.
CheckBit improved from 281.7 ns/op to 81.58 ns/op.

2. Reduction in Memory Usage (B/op):
BitPackingStorage:
SetBit reduced from 1200 B/op to 80 B/op.
CheckBit reduced from 399 B/op to 26 B/op.
ConventionalStorage:
SetBit reduced from 1200 B/op to 80 B/op.
CheckBit reduced from 427 B/op to 27 B/op.

3. Reduction in Allocations (allocs/op):
BitPackingStorage:
SetBit decreased from 30 allocs/op to 10 allocs/op.
CheckBit decreased from 9 allocs/op to 3 allocs/op.
ConventionalStorage:
SetBit decreased from 30 allocs/op to 10 allocs/op.
CheckBit decreased from 10 allocs/op to 3 allocs/op.
*/
// NumToBytes takes a numeric value and returns a slice of bytes representing that value in BigEndian format
/*func NumToBytes[T Hashable](num T) ([]byte, error) {
	buff := new(bytes.Buffer)
	switch k := any(num).(type) {
	case int:
		if err := binary.Write(buff, binary.BigEndian, int64(k)); err != nil {
			return nil, err
		}
	case uint:
		if err := binary.Write(buff, binary.BigEndian, uint64(k)); err != nil {
			return nil, err
		}
	case string:
		return []byte(k), nil
	case byte:
		return []byte{k}, nil
	case []byte:
		return k, nil
	default:
		if err := binary.Write(buff, binary.BigEndian, num); err != nil {
			return nil, err
		}
	}
	return buff.Bytes(), nil
}
*/
