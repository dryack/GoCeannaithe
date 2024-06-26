// file: toBytes_64bit.go
//go:build amd64 || arm64

package common

import (
	"fmt"
	"reflect"
	"testing"
)

// TODO:  TestNumToBytes still runs and passes - but I wasn't expecting that.  Need to investigate and be sure it's working correctly with the new TestNumToBytes

func TestNumToBytes(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected []byte
	}{
		{
			name:     "Test int",
			input:    int(300),
			expected: []byte{0, 0, 0, 0, 0, 0, 1, 44}, // Assuming a 64-bit architecture
		},
		{
			name:     "Test int32",
			input:    int32(300),
			expected: []byte{0, 0, 1, 44},
		},
		{
			name:     "Test float64",
			input:    float64(300.5),
			expected: []byte{64, 114, 200, 0, 0, 0, 0, 0},
		},
		// TODO: add more test cases
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reflect to handle the input type dynamically based on the type of tc.input
			result, err := callNumToBytesReflect(tc.input)
			if err != nil {
				t.Fatalf("Failed converting number to bytes: %v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test %s failed. Expected %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}

// callNumToBytesReflect uses reflection to call NumToBytes with the correct type parameter based on the runtime type of input.
func callNumToBytesReflect(input interface{}) ([]byte, error) {
	switch v := input.(type) {
	case int:
		return NumToBytes[int](v)
	case int32:
		return NumToBytes[int32](v)
	case float64:
		return NumToBytes[float64](v)
		// TODO: add more test cases
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}
