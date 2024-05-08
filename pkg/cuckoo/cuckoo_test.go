package cuckoo_test

import (
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"math/rand"
	"testing"

	"github.com/dryack/GoCeannaithe/pkg/cuckoo"
)

const (
	numBuckets = 2097152
	numKeys    = 1
)

var (
	keys     []int
	lookups  []int
	deletes  []int
	cf       *cuckoo.CuckooFilter[int]
	hashFunc = []uint8{common.Murmur3, common.Sha256, common.Sha512, common.SipHash, common.XXhash}
)

func init() {
	keys = generateKeys(numKeys)
	lookups = generateLookups(numKeys)
	deletes = generateDeletes(numKeys)
}

func BenchmarkCuckooFilter_Insert(b *testing.B) {
	for _, hash := range hashFunc {
		b.Run(fmt.Sprintf("Hash-%d", hash), func(b *testing.B) {
			cf = cuckoo.NewCuckooFilter[int](numBuckets)
			cf.WithHashFunction(hash)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, key := range keys {
					cf.Insert(key)
				}
			}
		})
	}
}

func BenchmarkCuckooFilter_Lookup(b *testing.B) {
	for _, hash := range hashFunc {
		b.Run(fmt.Sprintf("Hash-%d", hash), func(b *testing.B) {
			cf = cuckoo.NewCuckooFilter[int](numBuckets)
			cf.WithHashFunction(hash)
			for _, key := range keys {
				cf.Insert(key)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, key := range lookups {
					cf.Lookup(key)
				}
			}
		})
	}
}

func BenchmarkCuckooFilter_Delete(b *testing.B) {
	for _, hash := range hashFunc {
		b.Run(fmt.Sprintf("Hash-%d", hash), func(b *testing.B) {
			cf = cuckoo.NewCuckooFilter[int](numBuckets)
			cf.WithHashFunction(hash)
			for _, key := range keys {
				cf.Insert(key)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, key := range deletes {
					cf.Delete(key)
				}
			}
		})
	}
}

func generateKeys(n int) []int {
	keys := make([]int, n)
	for i := 0; i < n; i++ {
		keys[i] = rand.Int()
	}
	return keys
}

func generateLookups(n int) []int {
	lookups := make([]int, n)
	for i := 0; i < n; i++ {
		if rand.Float64() < 0.9 {
			lookups[i] = keys[rand.Intn(len(keys))]
		} else {
			lookups[i] = rand.Int()
		}
	}
	return lookups
}

func generateDeletes(n int) []int {
	deletes := make([]int, n)
	for i := 0; i < n; i++ {
		if rand.Float64() < 0.5 {
			deletes[i] = keys[rand.Intn(len(keys))]
		} else {
			deletes[i] = rand.Int()
		}
	}
	return deletes
}
