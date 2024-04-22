package main

import (
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/bloom"
	"log"
	"math/rand/v2"
)

func main() {
	var size uint64 = 1000000 // size of `bits` in the bloom filter, not the elements
	bf, _ := bloom.NewBloomFilter[string]().
		WithHashFunctions(7).
		WithStorage(bloom.NewBitPackingStorage[string](size, nil))

	err := bf.Storage.SetBit("test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bf.Storage.CheckBit("test")) // True
	fmt.Println(bf.Storage.CheckBit("nope")) // False
	// fmt.Println(bf.Storage)
	fmt.Println()

	bf2, _ := bloom.NewBloomFilter[int]().WithHashFunctions(5).WithStorage(bloom.NewBitPackingStorage[int](size, nil))
	err = bf2.Storage.SetBit(255)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(bf2.Storage.CheckBit(255)) // True
	fmt.Println(bf2.Storage.CheckBit(3))   // False
	fmt.Println()

	bf3, _ := bloom.NewBloomFilter[uint64]().WithHashFunctions(5).WithStorage(bloom.NewBitPackingStorage[uint64](size, nil))
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
	fmt.Println(bf4.Storage.CheckBit(2.71828)) // False
}
