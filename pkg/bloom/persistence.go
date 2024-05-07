package bloom

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"os"
	"reflect"
)

type Persistence[T common.Hashable] interface {
	Save(*BloomFilter[T]) error
	Load(*BloomFilter[T]) error
}

type FilePersistence[T common.Hashable] struct {
	directory string
	filename  string
}

func NewFilePersistence[T common.Hashable](directory, filename string) *FilePersistence[T] {
	return &FilePersistence[T]{directory: directory, filename: filename}
}

type BloomFilterData[T common.Hashable] struct {
	NumHashFunctions int
	Seeds            []uint32
	HashFunctionEnum uint8
	StorageData      []byte
	StorageType      string
	FilterType       string
}

func (bf *BloomFilter[T]) MarshalBinary() ([]byte, error) {
	gob.Register(&BloomFilterData[T]{})
	data := &BloomFilterData[T]{
		NumHashFunctions: bf.numHashFunctions,
		Seeds:            bf.seeds,
		HashFunctionEnum: bf.hashEnum,
		FilterType:       reflect.TypeOf(bf).String(),
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

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	encoder := gob.NewEncoder(gzipWriter)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (bf *BloomFilter[T]) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	var bfData BloomFilterData[T]
	decoder := gob.NewDecoder(gzipReader)
	if err := decoder.Decode(&bfData); err != nil {
		return err
	}

	if bfData.FilterType != reflect.TypeOf(bf).String() {
		return fmt.Errorf(
			"type mismatch: type during unmarshal (%s) doesn't match type during marshal (%s)",
			bfData.FilterType,
			reflect.TypeOf(bf).String(),
		)
	}

	bf.numHashFunctions = bfData.NumHashFunctions
	bf.seeds = bfData.Seeds

	switch bfData.HashFunctionEnum {
	case common.Murmur3:
		bf.hashFunction = common.HashKeyMurmur3
	case common.Sha256:
		bf.hashFunction = common.HashKeySha256
	case common.Sha512:
		bf.hashFunction = common.HashKeySha512
	case common.SipHash:
		bf.hashFunction = common.HashKeySipHash
	case common.XXhash:
		bf.hashFunction = common.HashKeyXXhash
	default:
		panic("unsupported hash function, this is probably a bug")
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

	tempfile, err := os.CreateTemp(fp.directory, fp.filename+"*")
	if err != nil {
		return err
	}

	_, err = tempfile.Write(data)
	if err != nil {
		return err
	}
	tempfile.Close() // manually close the tempfile so it can be renamed
	return os.Rename(tempfile.Name(), fp.getFullPath())
}

func (fp *FilePersistence[T]) Load(bf *BloomFilter[T]) error {
	data, err := os.ReadFile(fp.filename)
	if err != nil {
		return errors.New("error loading bloom filter: " + err.Error())
	}
	return bf.UnmarshalBinary(data)
}

func (fp *FilePersistence[T]) getFullPath() string {
	return fp.directory + string(os.PathSeparator) + fp.filename
}
