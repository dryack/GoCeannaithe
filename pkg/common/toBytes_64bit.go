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
