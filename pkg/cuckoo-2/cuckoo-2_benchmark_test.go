package cuckoo2

import (
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkFingerprint(b *testing.B) {
	bm, _ := NewBitManipulator(BitFieldConfig{
		FingerprintSize:  14,
		InUseSize:        1,
		FingerprintCount: 4,
	})
	rand.NewSource(time.Now().UnixNano())
	cf := &CuckooFilter[uint64]{}
	cf.bm = bm
	cf.WithHash(common.Murmur3)

	for i := 0; i < b.N; i++ {
		key := rand.Uint64()
		hash, _ := cf.hashFunction(key, 0)
		cf.fingerprint(hash)
	}
}

const (
	numSamples = 1_000_000
)

func TestFingerprintCollisions(t *testing.T) {
	bm, _ := NewBitManipulator(BitFieldConfig{
		FingerprintSize:  15,
		InUseSize:        1,
		FingerprintCount: 4,
	})
	rand.NewSource(time.Now().UnixNano())
	cf := &CuckooFilter[uint64]{}
	cf.bm = bm
	cf.WithHash(common.Murmur3)

	hashMap := make(map[Fingerprint]int)

	// Collect fingerprints
	for i := 0; i < numSamples; i++ {
		key := rand.Uint64()
		hash, _ := cf.hashFunction(key, 0)
		fp := cf.fingerprint(hash)
		hashMap[fp]++
	}

	uniqueCount := len(hashMap)
	collisionCount := 0

	for _, count := range hashMap {
		if count > 1 {
			collisionCount += count - 1
		}
	}

	uniquePercentage := (float64(uniqueCount) / float64(numSamples)) * 100
	collisionPercentage := (float64(collisionCount) / float64(numSamples)) * 100

	fmt.Printf("Total samples: %d\n", numSamples)
	fmt.Printf("Unique fingerprints: %d (%.2f%%)\n", uniqueCount, uniquePercentage)
	fmt.Printf("Collisions: %d (%.2f%%)\n", collisionCount, collisionPercentage)
}

func TestGetIndicesDistribution(t *testing.T) {
	const numSamples = 5000000
	const numBuckets = 10000000
	rand.NewSource(time.Now().UnixNano())
	bm, _ := NewBitManipulator(BitFieldConfig{
		FingerprintSize:  15,
		InUseSize:        1,
		FingerprintCount: 4,
	})
	cf := &CuckooFilter[uint64]{
		numBuckets: numBuckets,
		bm:         bm,
	}
	cf.WithHash(common.Murmur3)

	bucketCounts1 := make(map[uint32]int)
	bucketCounts2 := make(map[uint32]int)

	for i := 0; i < numSamples; i++ {
		hash := rand.Uint64()
		fp := Fingerprint(rand.Uint64())
		i1, i2 := cf.getIndices(hash, fp)
		bucketCounts1[i1]++
		bucketCounts2[i2]++
	}

	uniqueBuckets1 := len(bucketCounts1)
	uniqueBuckets2 := len(bucketCounts2)

	fmt.Printf("Total samples: %d\n", numSamples)
	fmt.Printf("Unique buckets i1: %d (%.2f%%)\n", uniqueBuckets1, (float64(uniqueBuckets1)/float64(numBuckets))*100)
	fmt.Printf("Unique buckets i2: %d (%.2f%%)\n", uniqueBuckets2, (float64(uniqueBuckets2)/float64(numBuckets))*100)
}

func BenchmarkGetIndicesDistribution(b *testing.B) {
	const numBuckets = 1000
	rand.NewSource(time.Now().UnixNano())
	bm, _ := NewBitManipulator(BitFieldConfig{
		FingerprintSize:  15,
		InUseSize:        1,
		FingerprintCount: 4,
	})
	cf := &CuckooFilter[uint64]{
		numBuckets: numBuckets,
		bm:         bm,
	}
	cf.WithHash(common.Murmur3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := rand.Uint64()
		fp := Fingerprint(rand.Uint64())
		cf.getIndices(hash, fp)
	}
}

func TestInsertEmptyBucket(t *testing.T) {
	config := BitFieldConfig{
		FingerprintSize:  15,
		InUseSize:        1,
		FingerprintCount: 4,
	}
	bm, err := NewBitManipulator(config)
	if err != nil {
		t.Fatalf("Failed to create BitManipulator: %v", err)
	}

	b := Bucket{}
	fp := Fingerprint(12345)

	success := b.insert(fp, *bm)
	if !success {
		t.Fatalf("Expected insert to succeed, but it failed")
	}

	expectedFP := bm.GetFingerprint(b.data, 0)
	if expectedFP != uint64(fp) {
		t.Errorf("Expected fingerprint %d at index 0, got %d", fp, expectedFP)
	}

	if !bm.IsInUse(b.data, 0) {
		t.Errorf("Expected index 0 to be in use")
	}
}

func TestInsertFullBucket(t *testing.T) {
	config := BitFieldConfig{
		FingerprintSize:  15,
		InUseSize:        1,
		FingerprintCount: 4,
	}
	bm, err := NewBitManipulator(config)
	if err != nil {
		t.Fatalf("Failed to create BitManipulator: %v", err)
	}

	b := Bucket{}
	for i := 0; uint32(i) < bm.FingerprintCount; i++ {
		fp := Fingerprint(i + 1)
		b.insert(fp, *bm)
	}

	fp := Fingerprint(12345)
	success := b.insert(fp, *bm)
	if success {
		t.Fatalf("Expected insert to fail, but it succeeded")
	}
}

func TestInsertPartiallyFilledBucket(t *testing.T) {
	config := BitFieldConfig{
		FingerprintSize:  15,
		InUseSize:        1,
		FingerprintCount: 4,
	}
	bm, err := NewBitManipulator(config)
	if err != nil {
		t.Fatalf("Failed to create BitManipulator: %v", err)
	}

	b := Bucket{}
	b.insert(Fingerprint(1), *bm) // Fill index 0
	b.insert(Fingerprint(2), *bm) // Fill index 1

	fp := Fingerprint(12345)
	success := b.insert(fp, *bm)
	if !success {
		t.Fatalf("Expected insert to succeed, but it failed")
	}

	expectedFP := bm.GetFingerprint(b.data, 2)
	if expectedFP != uint64(fp) {
		t.Errorf("Expected fingerprint %d at index 2, got %d", fp, expectedFP)
	}

	if !bm.IsInUse(b.data, 2) {
		t.Errorf("Expected index 2 to be in use")
	}
}

func TestInsertCuckooFilter(t *testing.T) {
	config := BitFieldConfig{
		FingerprintSize:  8,
		InUseSize:        1,
		FingerprintCount: 4,
	}
	bm, err := NewBitManipulator(config)
	if err != nil {
		t.Fatalf("Failed to create BitManipulator: %v", err)
	}

	// cf := NewCuckooFilter[uint64]()
	cf := &CuckooFilter[uint64]{
		buckets:    make([]*Bucket, 1024),
		numBuckets: 1024,
		bm:         bm,
		maxKicks:   1000,
	}
	cf.capacity = uint64(cf.numBuckets * cf.bm.FingerprintCount)

	for i := range cf.buckets {
		cf.buckets[i] = &Bucket{data: 0}
	}

	cf.WithHash(common.Murmur3)

	successfulInsertions := 0
	attemptedInsertions := 0

	// 1st insertion
	if !cf.Insert(12345) {
		t.Errorf("Expected insert to succeed")
		return
	}
	successfulInsertions++
	attemptedInsertions++

	numFingerprints := int(uint32(cf.bm.FingerprintCount) * cf.numBuckets)
	for i := 0; i <= numFingerprints; i++ {
		attemptedInsertions++
		if !cf.Insert(uint64(i)) {
			if cf.getLoadFactor() < 0.95 {
				t.Errorf("Expected insert to succeed: %d", i)
			}
		} else {
			successfulInsertions++
		}

	}
	// in general, we expect a partial-key cuckoo filter to survive a load factor of 95% when buckets > 4, and
	// fingerprint >= 7
	theoreticalFailureRate := big.NewFloat(.05)

	actualFailureRate := big.NewFloat(1).Sub(big.NewFloat(1), big.NewFloat(float64(successfulInsertions)/float64(attemptedInsertions)))
	actualFailureRate = actualFailureRate.Mul(actualFailureRate, big.NewFloat(100))
	fmt.Printf("failed insertion percentage: %s%%\n", actualFailureRate.Text('f', 3))
	fmt.Printf("theoretical insertion failure rate: %v%%\n", theoreticalFailureRate.Mul(theoreticalFailureRate, big.NewFloat(100)))
	fmt.Printf("inserted %.3f%% of keys\n", 100*float64(successfulInsertions)/float64(attemptedInsertions))

	// assert that actual failure rate is close to theoretical rate
	tolerance := big.NewFloat(0.01)
	difference := new(big.Float).Sub(actualFailureRate, theoreticalFailureRate)
	if difference.Cmp(tolerance) > 0 {
		t.Errorf("Actual failure rate (%v) deviates from theoretical rate (%v) by more than %v",
			actualFailureRate.Text('f', 6), theoreticalFailureRate.Text('f', 6), tolerance.Text('f', 2))
	}
}

// TODO: test for insertion failures as a percentage for various permutations of
// maxKicks and numBuckets - check against the theoretical expectations from the
// white paper
