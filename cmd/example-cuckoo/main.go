package main

import (
	"bufio"
	"github.com/dryack/GoCeannaithe/pkg/common"
	"github.com/dryack/GoCeannaithe/pkg/cuckoo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"os"
	"strconv"
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

	cuckoo2 := cuckoo.NewCuckooFilter[int64](2097152).WithHashFunction(common.Murmur3)
	p := message.NewPrinter(language.English)
	p.Printf("approximate size of cuckoo filter: %d bytes\n", cuckoo2.ApproximateSize())
	var numFails uint64

	file, _ := os.Open("./cmd/example-cuckoo/random_numbers.txt")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		x, _ := strconv.ParseInt(line, 10, 64)
		res := cuckoo2.Insert(x)
		if !res {
			numFails++
		}
	}
	p.Printf("number of failed insertions: %d\n", numFails)
}

/*
fp uint64:
approximate size of cuckoo filter: 100,663,308 bytes
number of failed insertions: 1,000,295

fp uint32:
approximate size of cuckoo filter: 67,108,876 bytes
number of failed insertions: 1,000,295

fp uint16:
approximate size of cuckoo filter: 50,331,660 bytes
number of failed insertions: 1,000,295

fp uint8:
approximate size of cuckoo filter: 41,943,052 bytes
number of failed insertions: 1,000,295

// after changes

fp uint8:
approximate size of cuckoo filter: 41,943,052 bytes
number of failed insertions: 999,987

fp uint16:
approximate size of cuckoo filter: 50,331,660 bytes
number of failed insertions: 999,987

fp uint32:
approximate size of cuckoo filter: 67,108,876 bytes
number of failed insertions: 999,987

fp uint64:
approximate size of cuckoo filter: 100,663,308 bytes
number of failed insertions: 999,987

*/
