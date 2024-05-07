// file: toBytes_32bit.go
//go:build 386 || arm

package common

import (
	"encoding/binary"
)

// NumToBytes takes a numeric value and returns a slice of bytes representing that value in BigEndian format
/*func NumToBytes[T Hashable](num T) ([]byte, error) {
	buff := new(bytes.Buffer)
	switch k := any(num).(type) {
	case int:
		if err := binary.Write(buff, binary.BigEndian, int32(k)); err != nil {
			return nil, err
		}
	case uint:
		if err := binary.Write(buff, binary.BigEndian, uint32(k)); err != nil {
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
}*/

func NumToBytes(num T) ([]byte, error) {
	switch k := num.(type) {
	case int:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint32(buf, uint64(k))
		return buf, nil
	case uint:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint32(buf, uint64(k))
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
