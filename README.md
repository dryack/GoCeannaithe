# GoCeannaithe
### (go kyan-eh-heh)
### Probabilistic Data Structures for Go projects

- [x] Bloom Filter
- [ ] Counting Bloom Filter
- [ ] Cuckoo Filter
- [ ] Golomb Compressed Set (midterm goal)
- [ ] Parallel-partitioned Bloom Filter (long term goal)
- [ ] Spatial Bloom Filter (long term goal)
- [ ] Layered Bloom Filter
- [ ] Count-min Sketch


## Bloom Filter Usage
### Manually configuring the bloom filter
When manually setting up a Bloom Filter, GoCeannaithe expects you to choose the number of bits in the filter (size),
and the number of hashes. While there's room for experimentation, in general there _is_ an optimal solution for a given
number of items you wish to store, and the probability of a false-positive you desire.  You can have a look at how the 
various parameters of a Bloom Filter interact, and get the optimal parameters, using the following calculator:
https://hur.st/bloomfilter/
```Go
import (
    "GoCeannaithe/pkg/bloom"
    "fmt"
    "log"
)

var size uint64 = 1000000 // size of `bits` in the bloom filter, not the elements
numHashes := 7 // number of different hashes to perform on each element
	
bf, log := bloom.NewBloomFilter[string]().
    WithHashFunctions(numHashes).
    WithStorage(bloom.NewBitPackingStorage[string](size, nil))
if err != nil {
log.Fatal(err)
}

err = bf.Storage.SetBit("Test")
if err != nil {
    log.Fatal(err)
}

fmt.Println(bf.Storage.CheckBit("monkey")) // True
fmt.Println(bf.Storage.CheckBit("nope")) // False
```
### Automatically configuring the Bloom Filter
Using WithAutoConfigure will utilize the best Storage Type and parameters for the number of elements you wish to store, 
and your desired rate of false positives (error rate)
```Go
import (
"GoCeannaithe/pkg/bloom"
"fmt"
"log"
)
size := 3000 // number of elements to be stored
errorRate := 0.10 // desired maximum error rate (10% in this case) 

bf4, err := bloom.NewBloomFilter[float32]().WithAutoConfigure(size, errorRate)
if err != nil {
    log.Fatal(err)
}
	
err = bf4.Storage.SetBit(3.14)
if err != nil {
   log.Fatal(err)
}
	
fmt.Println(bf4.Storage.CheckBit(3.14)) // True
fmt.Println(bf4.Storage.CheckBit(2.71828)) // False

```