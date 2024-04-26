package bloom

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"os"
	"reflect"
	"runtime"
)

type Persistence[T common.Hashable] interface {
	Save(*BloomFilter[T]) error
	Load(*BloomFilter[T]) error
}

type FilePersistence[T common.Hashable] struct {
	filepath string
}

func NewFilePersistence[T common.Hashable](filepath string) *FilePersistence[T] {
	return &FilePersistence[T]{filepath: filepath}
}

type BloomFilterData[T common.Hashable] struct {
	NumHashFunctions int
	Seeds            []uint32
	HashFunctionName string
	StorageData      []byte
	StorageType      string
}

func (bf *BloomFilter[T]) MarshalBinary() ([]byte, error) {
	data := &BloomFilterData[T]{
		NumHashFunctions: bf.numHashFunctions,
		Seeds:            bf.seeds,
		HashFunctionName: runtime.FuncForPC(reflect.ValueOf(bf.hashFunction).Pointer()).Name(),
	}

	switch storage := bf.Storage.(type) {
	case *BitPackingStorage[T]:
		data.StorageType = "BitPackingStorage"
		bits := make([]byte, len(storage.bits)*8)
		for i, v := range storage.bits {
			binary.LittleEndian.PutUint64(bits[i*8:], v)
		}
		data.StorageData = bits
	case *ConventionalStorage[T]:
		data.StorageType = "ConventionalStorage"
		bits := make([]byte, (len(storage.bits)+7)/8)
		for i, b := range storage.bits {
			if b {
				bits[i/8] |= 1 << (i % 8)
			}
		}
		data.StorageData = bits
	default:
		return nil, errors.New("unsupported storage type")
	}

	return json.Marshal(data)
}

func (bf *BloomFilter[T]) UnmarshalBinary(data []byte) error {
	var bfData BloomFilterData[T]
	err := json.Unmarshal(data, &bfData)
	if err != nil {
		return err
	}

	bf.numHashFunctions = bfData.NumHashFunctions
	bf.seeds = bfData.Seeds

	// TODO: really need to explore how we can better handle this, preferably without reflection. Probably just store a the hash used in BloomFilter as an int or sommething.
	switch bfData.HashFunctionName {
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeyMurmur":
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeyMurmur[...]":
		bf.hashFunction = common.HashKeyMurmur3[T]
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeySha256":
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeySha256[...]":
		bf.hashFunction = common.HashKeySha256[T]
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeySha512":
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeySha512[...]":
		bf.hashFunction = common.HashKeySha512[T]
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeySipHash":
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeySipHash[...]":
		bf.hashFunction = common.HashKeySipHash[T]
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeyXXhash":
	case "github.com/dryack/GoCeannaithe/pkg/common.HashKeyXXhash[...]":
	case "github.com/dryack/GoCeannaithe/pkg/bloom.(*BloomFilter[...]).UnmarshalBinary.func5":
		bf.hashFunction = common.HashKeyXXhash[T]
	case "github.com/dryack/GoCeannaithe/pkg/bloom.(*BloomFilter[...]).WithAutoConfigure.func1":
	case "github.com/dryack/GoCeannaithe/pkg/bloom.(*BloomFilter[...]).UnmarshalBinary.func6":
		bf.hashFunction = common.HashKeyMurmur3[T]
	default:
		return errors.New("unsupported hash function: " + bfData.HashFunctionName)
	}

	switch bfData.StorageType {
	case "BitPackingStorage":
		bits := make([]uint64, len(bfData.StorageData)/8)
		for i := range bits {
			bits[i] = binary.LittleEndian.Uint64(bfData.StorageData[i*8:])
		}
		bf.Storage = &BitPackingStorage[T]{
			bits:        bits,
			seeds:       bf.seeds,
			bitsLength:  uint64(len(bits)),
			bloomFilter: bf,
		}
	case "ConventionalStorage":
		bits := make([]bool, len(bfData.StorageData)*8)
		for i := range bits {
			bits[i] = bfData.StorageData[i/8]&(1<<(i%8)) != 0
		}
		bf.Storage = &ConventionalStorage[T]{
			bits:        bits,
			seeds:       bf.seeds,
			sliceLength: uint64(len(bits)),
			bloomFilter: bf,
		}
	default:
		return errors.New("unsupported storage type")
	}

	return nil
}

func (fp *FilePersistence[T]) Save(bf *BloomFilter[T]) error {
	data, err := bf.MarshalBinary()
	if err != nil {
		return err
	}
	return os.WriteFile(fp.filepath, data, 0644)
}

func (fp *FilePersistence[T]) Load(bf *BloomFilter[T]) error {
	data, err := os.ReadFile(fp.filepath)
	if err != nil {
		return err
	}
	return bf.UnmarshalBinary(data)
}
