package cuckoo2

import "fmt"

type BitField uint64

type BitFieldConfig struct {
	FingerprintSize  int
	InUseSize        int
	FingerprintCount uint32
}

type BitManipulator struct {
	BitFieldConfig
	TotalBits            int
	FingerprintPositions []int
	InUsePositions       []int
	CounterPositions     []int
	FingerprintMask      uint64
	InUseMask            uint64
	CounterMask          uint64
	CounterMax           uint32
}

func NewBitManipulator(config BitFieldConfig) (*BitManipulator, error) {
	bm := &BitManipulator{
		BitFieldConfig: config,
	}

	// Validate and calculate positions
	bitPos := 0
	for i := 0; uint32(i) < bm.FingerprintCount; i++ {
		if bitPos+bm.FingerprintSize > 64 {
			return nil, fmt.Errorf("not enough bits for the given configuration")
		}
		bm.FingerprintPositions = append(bm.FingerprintPositions, bitPos)
		bitPos += bm.FingerprintSize

		if bitPos+bm.InUseSize > 64 {
			return nil, fmt.Errorf("not enough bits for the given configuration")
		}
		bm.InUsePositions = append(bm.InUsePositions, bitPos)
		bitPos += bm.InUseSize

		bm.CounterPositions = append(bm.CounterPositions, bitPos)
	}

	bm.TotalBits = bitPos

	// Calculate masks
	bm.FingerprintMask = (1 << bm.FingerprintSize) - 1
	bm.InUseMask = (1 << bm.InUseSize) - 1

	return bm, nil
}

func (bm *BitManipulator) SetFingerprint(bf *BitField, index int, value uint64) {
	pos := bm.FingerprintPositions[index]
	*bf = BitField((uint64(*bf) & ^(bm.FingerprintMask << pos)) | ((value & bm.FingerprintMask) << pos))
}

func (bm *BitManipulator) GetFingerprint(bf BitField, index int) uint64 {
	pos := bm.FingerprintPositions[index]
	return (uint64(bf) >> pos) & bm.FingerprintMask
}

func (bm *BitManipulator) SetInUse(bf *BitField, index int, inUse bool) {
	pos := bm.InUsePositions[index]
	if inUse {
		*bf |= BitField(bm.InUseMask << pos)
	} else {
		*bf &^= BitField(bm.InUseMask << pos)
	}
}

func (bm *BitManipulator) IsInUse(bf BitField, index int) bool {
	pos := bm.InUsePositions[index]
	return (uint64(bf) & (bm.InUseMask << pos)) != 0
}

func (bm *BitManipulator) SetCounter(bf *BitField, index int, value uint64) {
	pos := bm.CounterPositions[index]
	*bf = BitField((uint64(*bf) & ^(bm.CounterMask << pos)) | ((value & bm.CounterMask) << pos))
}

func (bm *BitManipulator) IncrementCounter(bf *BitField, index int) {
	currentValue := bm.GetCounter(*bf, index)
	/*if currentValue+1 >= bm.GetCounterMax(*bf) {
		return
	}*/
	fmt.Println("incrementing counter") // DEBUG
	pos := bm.CounterPositions[index]
	*bf = BitField((uint64(*bf) & ^(bm.CounterMask << pos)) | ((currentValue + 1&bm.CounterMask) << pos))
}

func (bm *BitManipulator) GetCounter(bf BitField, index int) uint64 {
	pos := bm.CounterPositions[index]
	return (uint64(bf) >> pos) & bm.CounterMask
}

func (bm *BitManipulator) GetCounterMax(bf BitField) uint64 {
	return bm.CounterMask
}
