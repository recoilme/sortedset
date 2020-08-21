package ordset

import (
	"fmt"
	"math/rand"
	"testing"
)

var rnd *rand.Rand

func init() {
	seed := int64(1597343679089175000) //int64(time.Now().UnixNano())
	fmt.Printf("seed: %d\n", seed)
	rnd = rand.New(rand.NewSource(seed))
}

func randKeysBin(N int) (keys [][]byte) {
	n := 8
	for i := 0; i < N; i++ {
		s := make([]byte, n)
		rnd.Read(s)
		for i := 0; i < n; i++ {
			s[i] = 'a' + (s[i] % 26)
		}
		keys = append(keys, (s))
	}
	return
}

/*
func randKeys(N int) (keys []string) {
	format := fmt.Sprintf("%%0%dd", len(fmt.Sprintf("%d", N-1)))
	for _, i := range rand.Perm(N) {
		keys = append(keys, fmt.Sprintf(format, i))
	}
	return
}

func stringsEquals(a, b []string) (int, bool) {
	if len(a) != len(b) {
		return -1, false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return i, false
		}
	}
	return 0, true
}


//go test -benchmem -bench Add

func BenchmarkAddAsc(b *testing.B) {
	keys := randKeys(b.N)
	sort.Strings(keys)
	set := New(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}



func BenchmarkAddAscBin(b *testing.B) {
	keys := randKeysBin(b.N)
	sort.Strings(keys)
	set := New(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}

func BenchmarkAddDesc(b *testing.B) {
	keys := randKeys(b.N)
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	set := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}

func BenchmarkAddDescBin(b *testing.B) {
	keys := randKeysBin(b.N)
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	set := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}

func BenchmarkAddRand(b *testing.B) {
	keys := randKeys(b.N)
	set := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}*/

func BenchmarkAddRandBin(b *testing.B) {
	keys := randKeysBin(b.N)
	set := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}

/*
func BenchmarkParallel(b *testing.B) {
	set := New()
	keys := randKeysBin(b.N)
	var i uint64 = 0
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			set.Put(keys[atomic.LoadUint64(&i)])
			atomic.AddUint64(&i, 1)
		}
	})
}

func BenchmarkKeys(b *testing.B) {
	keys := randKeysBin(b.N)
	set := New()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
	b.ResetTimer()
	for _, key := range set.Keys() {
		_ = key
	}
}

func BenchmarkHas(b *testing.B) {
	keys := randKeysBin(b.N)
	set := New()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
	b.ResetTimer()
	for _, key := range keys {
		h := set.Has(key)
		if !h {
			b.Fatal("bad news")
		}
	}
}

func BenchmarkSearchBin(b *testing.B) {
	keys := randKeysBin(b.N)
	keysSorted := make([]string, b.N)
	set := New()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
		keysSorted[i] = keys[i]
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keysSorted)))
	b.ResetTimer()
	for _, key := range set.Keys() {
		sort.Search(b.N, func(n int) bool {
			return keysSorted[n] <= key
		})
	}
}

func BenchmarkSearchNewBin(b *testing.B) {
	keys := randKeysBin(b.N)
	keysSorted := make([]string, b.N)
	set := New()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
		keysSorted[i] = keys[i]
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keysSorted)))
	b.ResetTimer()
	for _, key := range set.Keys() {
		SearchBinary(0, b.N, func(n int) bool {
			return keysSorted[n] <= key
		})
	}
}

func TestSearchBin(t *testing.T) {
	N := 10000
	keys := randKeysBin(N)
	keysSorted := make([]string, N)
	for i := 0; i < N; i++ {
		keysSorted[i] = keys[i]
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keysSorted)))
	find := keys[N/2]
	i := SearchBinary(0, N, func(n int) bool {
		return keysSorted[n] <= find
	})
	fmt.Println("ind", i)
	assert.Equal(t, find, keysSorted[i])
}

func SearchBinary(i, j int, f func(int) bool) int {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	//bound := 1
	//for bound < j && f(j-1) {
	//	bound *= 2
	//}
	nh := i
	if j > 10000 {
		if i < j {
			h := i
			for h < j && !f(h) {
				nh = h
				h = int(uint(h+j) >> 1)
			}
		}
		//fmt.Println("new h", nh)
		i = nh
	}
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if !f(h) {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}

func TestDescend(t *testing.T) {
	set := New()
	var all []string
	for i := 512; i >= 0; i-- {
		var key string
		key = fmt.Sprintf("%03d", i)
		all = append(all, key)
		set.Put(key)
	}
	assert.Equal(t, len(all), len(set.Keys()))

	ind, eq := stringsEquals(set.Keys(), all)
	if !eq {
		fmt.Printf("ind: %v\n", ind)
		t.Fatal("mismatch")
	}
}

func TestRand(t *testing.T) {
	for i := 3; i < 3333; i = i + 112 {
		testRand(i, t)
	}
}

func testRand(N int, t *testing.T) {

	keys := randKeysBin(N)
	set := New()
	for _, key := range keys {
		set.Put(key)
	}
	result := set.Keys()
	//set.print()

	assert.Equal(t, len(keys), len(result))
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	ind, eq := stringsEquals(result, keys)
	if !eq {
		fmt.Printf("ind: %v\n", ind)
		fmt.Println(keys)
		fmt.Println(result)
		t.Fatal("mismatch")
	}
}
func TestSearch(t *testing.T) {
	//prepend
	keys := [6]string{"5", "3", "1", "", "", ""}
	i := sort.Search(len(keys), func(i int) bool {
		return keys[i] <= "6"
	})
	assert.Equal(t, 0, i)
	n := 3
	copy(keys[i+1:n+1], keys[i:n])
	keys[i] = "6"

	//insert
	//fmt.Println(keys) //[6 5 3 1  ]
	i = sort.Search(len(keys), func(i int) bool {
		return keys[i] <= "4"
	})
	assert.Equal(t, 2, i)
	n = 4
	copy(keys[i+1:n+1], keys[i:n])
	keys[i] = "4"

	//append
	//fmt.Println(keys) //[6 5 4 3 1  ]
	i = sort.Search(len(keys), func(i int) bool {
		return keys[i] <= "0"
	})
	assert.Equal(t, 5, i)
	n = 5
	keys[i] = "0"
	//fmt.Println(keys)
	//[6 5 4 3 1 0]

	keys = [6]string{"", "", "", "", "", ""}
	i = sort.Search(len(keys), func(i int) bool {
		return keys[i] <= "0"
	})
	assert.Equal(t, 0, i)
}

func TestSliceIns(t *testing.T) {
	slice := []string{"xgvfrjpr", "lyqlvxfg", "lvzaatri", "dcwowvga", "csslcwvn", "ajfwsdnf"}

	idx := 0
	slice = append(slice, "_", "_")
	copy(slice[idx*2+2:], slice[idx*2+1:])
	slice[idx*2+1] = "6"
	//fmt.Printf("%+v\n", slice)
	//[xgvfrjpr 6 lyqlvxfg lvzaatri dcwowvga csslcwvn ajfwsdnf]
	copy(slice[idx*2+3:], slice[idx*2+2:])
	slice[idx*2+2] = "5"
	//fmt.Printf("%+v\n", slice)
	//[xgvfrjpr 6 5 lyqlvxfg lvzaatri dcwowvga csslcwvn ajfwsdnf]

	slice = []string{"xgvfrjpr", "lyqlvxfg", "lvzaatri", "dcwowvga", "csslcwvn", "ajfwsdnf"}

	idx = 0
	slice = append(slice, "_", "_")
	copy(slice[idx*2+3:], slice[idx*2+1:])
	slice[idx*2+1] = "6"
	slice[idx*2+2] = "5"
	//fmt.Printf("%+v\n", slice)
}

func TestAscend(t *testing.T) {
	set := New()
	var all []string
	for i := 0; i < 1126; i++ {
		var key string
		key = fmt.Sprintf("%04d", i)
		all = append(all, key)
		set.Put(key)
	}
	res := set.Keys()
	sort.Sort(sort.StringSlice(res))
	assert.Equal(t, len(all), len(res))

	ind, eq := stringsEquals(res, all)
	if !eq {
		fmt.Printf("ind: %v\n", ind)
		fmt.Println(all)
		fmt.Println(res)
		t.Fatal("mismatch")
	}
}

func TestParallel(t *testing.T) {
	N := 10000
	set := New()
	keys := randKeys(N)
	bkt := Bucket(set, "")
	lotsa.Output = nil
	lotsa.MemUsage = true
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		bkt.Put(keys[i])
	})
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	assert.Equal(t, keys, bkt.Keys())
}

func TestBucket(t *testing.T) {
	set := New()
	bkt := Bucket(set, "user")
	bkt.Put("01")
	assert.Equal(t, []string{"user01"}, set.Keys())
	users := Bucket(set, "user")
	users.Put("rob")
	users.Put("bob")
	users.Put("pike")
	users.Put("alice")
	users.Put("anna")
	items := Bucket(set, "item")

	keys := randKeys(260)
	for _, key := range keys {
		items.Put(key)
	}
	items.Put("003")
	set.Put("item") //"itemitem"

	assert.Equal(t, "rob", users.Cursor().Last())
	assert.Equal(t, "259", items.Cursor().Last())

	c := users.Cursor()
	var first string
	for k := c.Last(); k != ""; k = c.Prev() {
		first = k
	}
	assert.Equal(t, "01", first)

	c = items.Cursor()
	first = ""
	for k := c.Last(); k != ""; k = c.Prev() {
		first = k
	}
	assert.Equal(t, "000", first)
}

func TestCursor(t *testing.T) {
	set := New()
	keys := randKeys(7)
	bkt := Bucket(set, "")
	for _, key := range keys {
		bkt.Put(key)
	}
	c := bkt.Cursor()
	assert.Equal(t, "6", c.Last())
	//descend

	for k := c.Last(); k != ""; k = c.Prev() {
		_ = k
		fmt.Printf("[%s] ", k)
		//set.Delete(k)
	}
}

func TestDelete(t *testing.T) {
	set := New()
	keys := randKeys(7)
	bkt := Bucket(set, "")
	for _, key := range keys {
		bkt.Put(key)
	}
	c := bkt.Cursor()
	assert.Equal(t, "6", c.Last())
	//descend
	set.Delete("0")
	set.Delete("1")
	set.Delete("4")
	set.Delete("6")
	assert.Equal(t, "5", c.Last())
	set.Delete("5")
	set.Delete("2")
	assert.Equal(t, "3", c.Last())
	set.Delete("3")
	assert.Equal(t, "", c.Last())
}
*/
/*
BenchmarkAddAsc-8                3545025               574 ns/op              38 B/op          0 allocs/op
BenchmarkAddAscBin-8             2262948               563 ns/op              38 B/op          0 allocs/op
BenchmarkAddDesc-8               3241903               484 ns/op              38 B/op          0 allocs/op
BenchmarkAddDescBin-8            2993749               473 ns/op              38 B/op          0 allocs/op
BenchmarkAddRand-8               1000000              1081 ns/op              27 B/op          0 allocs/op
BenchmarkAddRandBin-8            1000000              1052 ns/op              27 B/op          0 allocs/op
BenchmarkParallel-8              1000000              1034 ns/op              27 B/op          0 allocs/op
*/
