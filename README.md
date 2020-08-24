# `sortedset`

[![GoDoc](https://godoc.org/github.com/recoilme/sortedset?status.svg)](https://godoc.org/github.com/recoilme/sortedset)

Package sortedset provide sorted set, with strings/binary comparator, backed by arrays

## Status

Code review

## Usage

install Go and run ```go get```:

```go
go get github.com/recoilme/sortedset
```

## Motivation

Set's usualy based on Trees. Trees is:

- based on pointers => many allocations
- based on pointers => memory fragmentation
- rebalansed on the fly => perfomance degradation, or not safe for traversting
- degradation with grow

## Architecture

`sortedset` is based on custom data structure. Data stored ih fixed size arrays, pages: ```[256]string``` with pages indexes (max/min). Then you put key, `sortedset` will store data in descending order with binary comparator.

If pageSize = 4, and insert ```"1","3","5","7","9"``` -  data will look's like:

```go
index: "9","3","1","1"
pages: page0["9", "7", "5", "3"] page1["1","","",""]
```

On overflow - page split on the mid.
On insert: 
 - scan index with binary search
 - scan page with binary search
 - insert

Let insert "6":

```go
index: "9","6","5","3","1","1"
pages: page0["9", "6", "", ""] page1["5","3","",""] page2["1","","",""]
```

That's all. 

### New/Put

Put is **safe** for use from multiple goroutines.

```go
	set := sortedset.New()
	set.Put("a")
	set.Put("b")
	fmt.Println(set.Keys())
	//[b a]
```

### Buckets

Buckets are keys with same prefix. Methods of buckets are **safe** for concurrent usage.

```go
	set := sortedset.New()
	users := sortedset.Bucket(set, "user")//Bucket name may be ommited
	users.Put("rob")
	users.Put("bob")
	users.Put("pike")
	users.Put("alice")
	fmt.Println(users.Keys(0,0))
	// output: [rob pike bob alice]
    
	items := sortedset.Bucket(set, "item")
	items.Put("003")
	items.Put("042")
	fmt.Println(items.Keys(0,0))
	// output: [042 003]
```

### Iterating over keys

`sortedset` stores its keys in byte-sorted descending order. This makes sequential iteration over these keys extremely fast. To iterate over keys we'll use a `Cursor`:

```go
	fmt.Println("Cursor")
	set := sortedset.New()
	users := sortedset.Bucket(set, "user")
	users.Put("rob")
	users.Put("bob")
	users.Put("pike")
	users.Put("alice")
	users.Put("anna")
	items := sortedset.Bucket(set, "item")
	items.Put("003")
	c := users.Cursor()
	for k := c.Last(); k != ""; k = c.Prev() {
		fmt.Printf("[%s] ", k)
	}
	fmt.Println()
	//[rob] [pike] [bob] [anna] [alice]

	c = items.Cursor()
	for k := c.Last(); k != ""; k = c.Prev() {
		fmt.Printf("[%s] ", k)
	}
	fmt.Println()
	//[003]
```

The cursor allows you to move to a specific point in the list of keys and move backward through the keys one at a time.

The following functions are available on the cursor:

```go
Last()   Move to the last key.
Prev()   Move to the previous key.
```

You must seek to a position using Last() before calling Prev(). If you do not seek to a position then these functions will return a empty key.

Cursor is method of bucket and safe for concurrent usage. Data in cursor are must no panic but if underlaing array is modified, result will be unexpected.

### Benchmark

**BenchmarkParallel:**
```
Put: 
  100,000 ops over 8 threads in 49ms, 2,059,360/sec, 485 ns/op, 2.5 MB, 25 bytes/op
1,000,000 ops over 8 threads in 1041ms, 960,693/sec, 1040 ns/op, 26.4 MB, 27 bytes/op
```

**BenchmarkSequental:**
```
goos: darwin
goarch: amd64
pkg: github.com/recoilme/sortedset
BenchmarkKeys-8                18761932                87.6 ns/op            86 B/op          0 allocs/op
BenchmarkAddAsc-8                3533504               388 ns/op              38 B/op          0 allocs/op
BenchmarkAddAscBin-8             2423774               521 ns/op              38 B/op          0 allocs/op
BenchmarkAddDesc-8               3155502               371 ns/op              38 B/op          0 allocs/op
BenchmarkAddDescBin-8            3435102               524 ns/op              38 B/op          0 allocs/op
BenchmarkAddRand-8               1000000              1051 ns/op              27 B/op          0 allocs/op
BenchmarkAddRandBin-8            1000000              1033 ns/op              27 B/op          0 allocs/op
BenchmarkParallel-8              1000000              1018 ns/op              27 B/op          0 allocs/op
BenchmarkHas-8           	 1000000              1035 ns/op               0 B/op          0 allocs/op
```

**Left-Leaning Red-Black (LLRB) implementation of 2-3 balanced binary search trees**
[github.com/google/btree](github.com/google/btree)

```
OneThreadOrdSet
10,000,000 ops in 22245ms, 449,547/sec, 2224 ns/op, 273.2 MB, 28 bytes/op
OneThreadGoogle
10,000,000 ops in 29001ms, 344,814/sec, 2900 ns/op, 431.2 MB, 45 bytes/op
ParallelOrdSet
10,000,000 ops over 8 threads in 21492ms, 465,289/sec, 2149 ns/op, 266.1 MB, 27 bytes/op
```

### TODO

 - seek()
 - first()
 - next()

## Contact

Vadim Kulibaba [@recoilme](http://t.me/recoilme)

## License

`sortedset` source code is available under the MIT [License](/LICENSE).
