package bloom

import (
	"testing"
)

func createSeeds(n int) []uint32 {
	seeds := make([]uint32, n)
	for i := range seeds {
		seeds[i] = uint32(i + 1)
	}
	return seeds
}

// Benchmark for BitPackingStorage SetBit
func BenchmarkBitPackingStorageSetBit(b *testing.B) {
	seeds := createSeeds(10)
	storage := NewBitPackingStorage[int](1000000, seeds)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.SetBit(i % 1000000)
	}
}

// Benchmark for ConventionalStorage SetBit
func BenchmarkConventionalStorageSetBit(b *testing.B) {
	seeds := createSeeds(10)
	storage := NewConventionalStorage[int](1000000, seeds)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.SetBit(i % 1000000)
	}
}

// Benchmark for BitPackingStorage CheckBit
func BenchmarkBitPackingStorageCheckBit(b *testing.B) {
	seeds := createSeeds(10)
	storage := NewBitPackingStorage[int](1000000, seeds)
	// Setting some bits
	for i := 0; i < 100000; i++ {
		storage.SetBit(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.CheckBit(i % 1000000)
	}
}

// Benchmark for ConventionalStorage CheckBit
func BenchmarkConventionalStorageCheckBit(b *testing.B) {
	seeds := createSeeds(10)
	storage := NewConventionalStorage[int](1000000, seeds)
	// Setting some bits
	for i := 0; i < 100000; i++ {
		storage.SetBit(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.CheckBit(i % 1000000)
	}
}
