package main

import (
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/bloom"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"log"
	"math/rand/v2"
)

func main() {
	/*var size uint64 = 1000000 // size of `bits` in the bloom filter, not the elements
	bf, _ := bloom.NewBloomFilter[string]().
		WithHashFunctions(7, common.HashKeyMurmur3[string]).
		WithStorage(bloom.NewBitPackingStorage[string](size, nil))

	err := bf.Storage.SetBit("test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bf.Storage.CheckBit("test")) // True
	fmt.Println(bf.Storage.CheckBit("nope")) // False
	// fmt.Println(bf.Storage)
	fmt.Println()

	bf2, _ := bloom.NewBloomFilter[int]().WithHashFunctions(5, common.HashKeySipHash[int]).WithStorage(bloom.NewBitPackingStorage[int](size, nil))
	err = bf2.Storage.SetBit(255)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bf2.Storage.CheckBit(255)) // True
	fmt.Println(bf2.Storage.CheckBit(3))   // False
	fmt.Println()

	bf3, _ := bloom.NewBloomFilter[uint64]().WithHashFunctions(5, common.HashKeyXXhash[uint64]).WithStorage(bloom.NewBitPackingStorage[uint64](size, nil))
	err = bf3.Storage.SetBit(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bf3.Storage.CheckBit(2)) // True
	fmt.Println(bf3.Storage.CheckBit(3)) // False
	fmt.Println()

	size = 3000 //
	// bf4, _ := bloom.NewBloomFilter[float32]().WithHashFunctions(3).WithStorage(bloom.NewConventionalStorage[float32](size, nil))
	bf4, err := bloom.NewBloomFilter[float32]().WithAutoConfigure(size, 0.10)
	if err != nil {
		log.Fatal(err)
	}
	err = bf4.Storage.SetBit(3.14)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 500; i++ {
		x := rand.Float32()
		err = bf4.Storage.SetBit(x)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(bf4.Storage.CheckBit(3.14))    // True
	fmt.Println(bf4.Storage.CheckBit(2.71828)) // False*/

	size := uint64(10_000_000)
	// errorRate := 0.015
	// bf5, _ := bloom.NewBloomFilter[int]().WithPersistence(bloom.NewFilePersistence[int]("bf_data.dat")).WithAutoConfigure(size, errorRate)
	bf5, _ := bloom.NewBloomFilter[int]().WithHashFunctions(5, common.XXhash).WithPersistence(bloom.NewFilePersistence[int]("bf_data.dat")).WithStorage(bloom.NewBitPackingStorage[int](size, nil))
	for i := 0; i < 100; i++ {
		bf5.Storage.SetBit(i)
	}
	err := bf5.SavePersistence()
	if err != nil {
		log.Fatal(err)
	}

	bf6 := bloom.NewBloomFilter[int]().WithPersistence(bloom.NewFilePersistence[int]("bf_data.dat"))
	err = bf6.LoadPersistence()
	if err != nil {
		fmt.Println("Error loading Bloom filter:", err)
		return
	}
	fmt.Println(bf6.Storage.CheckBit(50))
	fmt.Println(bf6.Storage.CheckBit(150))
	if !bf6.Storage.CheckBit(200) {
		bf6.Storage.SetBit(200)
	} else {
		bf6.Storage.SetBit(rand.Int())
	}
	err = bf6.SavePersistence()
	if err != nil {
		fmt.Println("Error saving Bloom filter:", err)
		return
	}

}
