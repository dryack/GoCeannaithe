package main

import (
	"fmt"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"github.com/dryack/GoCeannaithe/pkg/cuckoo"
	"math/rand"
)

func main() {
	/*fmt.Println("cuckoo:")
	cf := cuckoo.NewCuckooFilter[int](2048).WithHashFunction(common.Murmur3)
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("lookup:", cf.Lookup(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("insert:", cf.Insert(5))
	fmt.Println("lookup (expect false):", cf.Lookup(6))
	fmt.Println("delete (expect false):", cf.Delete(6))
	fmt.Println("delete (expect true)", cf.Delete(5))

	fmt.Println("cuckoo w/strings:")
	cf2 := cuckoo.NewCuckooFilter[string](2048).WithHashFunction(common.Murmur3)
	cf2.Insert("a duck")
	cf2.Insert("a duct")
	cf2.Insert("a duck")
	cf2.Insert("a duck")
	cf2.Insert("a duck")
	cf2.Insert("a duck")*/

	/*cucko1 := cuckoo.NewCuckooFilter[int](1024).WithHashFunction(common.Murmur3)
	for i := 0; i < 50; i++ {
		fmt.Println(cucko1.Insert(5))
	}*/

	cucko2 := cuckoo.NewCuckooFilter[int](32768).WithHashFunction(common.Murmur3)
	for i := 0; i < 5000; i++ {
		x := rand.Intn(1000)
		fmt.Println(x, cucko2.Insert(x))
	}
}
