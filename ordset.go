// Package ordset provide ordered set, with strings comparator, backed by arrays
package ordset

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const pageSize = 256

type page struct {
	items    [pageSize]string
	min      string
	max      string
	numItems int
	modified bool
}

// OrdSet provide ordered set, with strings comparator
type OrdSet struct {
	idxs  []string
	pages []*page
}

type SyncBucket struct {
	sync.RWMutex
	Name    string
	Set     *OrdSet
	idxPage int
	idxItem int
}

type Cursor struct {
	bucket *SyncBucket
}

// New create ordered set with capacity (first param),
// default is 1024, must be > 3 and power of 2
func New(intParams ...int) *OrdSet {
	capacity := 1024
	if len(intParams) > 0 && intParams[0] > 4 {
		capacity = int(nextPowerOf2(uint32(intParams[0])))
	}
	p := &page{}
	set := &OrdSet{}
	set.pages = make([]*page, 0, capacity)
	set.idxs = make([]string, 0, capacity*2)
	set.pages = append(set.pages, p)
	set.idxs = append(set.idxs, "", "")
	return set
}

// Put will add key in set, if not present
// Bucket may be empty
func (set *OrdSet) Put(key string) {
	//fmt.Printf("Add %s %+v\n", key, set)
	// sort desc
	i := sort.Search(len(set.idxs), func(n int) bool {
		return set.idxs[n] <= key
	})
	if i < len(set.idxs) && set.idxs[i] == key {
		// key is present at data[i], nothing to do here
		return
	}

	idx := i / 2
	if i == len(set.idxs) {
		//not found - append to last
		idx = len(set.pages) - 1
	}
	p := set.pages[idx]
	if p.numItems == pageSize-1 {
		set.split(idx)
		set.Put(key)
		return
	}
	p = p.add(key)
	//fmt.Printf("set before:%+v\n", set)
	if p.modified {
		// update indexes
		set.idxs[idx*2] = p.max
		set.idxs[idx*2+1] = p.min
	}
	//fmt.Printf("Ret %s %+v\n", key, set)
}

func (p *page) add(key string) *page {
	p.modified = false
	//fmt.Println("add", key)
	// desc
	i := sort.Search(p.numItems, func(n int) bool {
		return p.items[n] <= key
	})
	//fmt.Println("page i", i, key, p.items[1] == key)
	if i < p.numItems && p.items[i] == key {
		// key is present at data[i], nothing to do here
		return p
	}
	if i == p.numItems {
		// not found, new min, append at the end
		p.items[i] = key
		p.max = p.items[0]
		p.min = key
		p.numItems++
		p.modified = true
		//fmt.Println("data i == p.numItems:", p.items, p.min, p.max, p.numItems)
		return p
	}

	//insert or prepend
	if i == 0 {
		//prepend, new max
		p.max = key
		p.min = p.items[p.numItems-1]
		p.modified = true
	}
	//insert - not modify min/max
	copy(p.items[i+1:p.numItems+1], p.items[i:p.numItems])
	p.items[i] = key
	p.numItems++
	return p
}

func (set *OrdSet) split(idx int) {
	//fmt.Printf("set before split:%+v\n", set)
	//example data: 015 014 013 012 011 010 009 008 007 006 005...
	p := set.pages[idx]
	//fmt.Println("data before:", p.items, p.min, p.max, p.numItems)
	mid := (pageSize - 1) / 2 //127
	pRight := &page{}
	copy(pRight.items[:mid+1], p.items[mid:])
	//0:126 127:254
	//right
	pRight.max = p.items[mid]          //127
	pRight.min = p.items[p.numItems-1] //254
	pRight.numItems = mid + 1          //128
	//left
	p.numItems = mid //254 -> 127
	p.max = p.items[0]
	p.min = p.items[mid-1] //[126]
	//for i := mid; i < pageSize; i++ {
	//	p.items[i] = ""
	//}
	//grow pages
	set.pages = append(set.pages, nil)
	//copy
	copy(set.pages[idx+1:], set.pages[idx:])
	set.pages[idx] = p
	set.pages[idx+1] = pRight
	//grow idx on 2
	set.idxs = append(set.idxs, "", "")
	copy(set.idxs[idx*2+3:], set.idxs[idx*2+1:])

	set.idxs[idx*2+1] = p.min
	set.idxs[idx*2+2] = pRight.max
	/*
		fmt.Println("data left:", p.items, p.min, p.max, p.numItems)
		fmt.Println("data right:", pRight.items, pRight.min, pRight.max, pRight.numItems)
		fmt.Printf("set after split: idx:%d lenidx:%d set %+v %d\n", idx, len(set.idxs), set, len(set.idxs))
	*/
}

// Keys return all keys
func (set *OrdSet) Keys() (result []string) {
	for _, p := range set.pages {
		for i, key := range p.items {
			if i >= p.numItems {
				break
			}
			result = append(result, key)
		}
	}
	return result
}

func (set *OrdSet) print() (result []string) {
	for i, p := range set.pages {
		fmt.Printf("i:%d max:%s min:%s\n", i, p.max, p.min)
	}
	for i, idx := range set.idxs {
		fmt.Printf("i:%d max:%s\n", i, idx)
	}
	return result
}

// https://github.com/thejerf/gomempool/blob/master/pool.go#L519
// http://graphics.stanford.edu/~seander/bithacks.html#RoundUpPowerOf2
// suitably modified to work on 32-bit
func nextPowerOf2(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++

	return v
}

func Bucket(set *OrdSet, name string) *SyncBucket {
	return &SyncBucket{Name: name, Set: set}
}

// Put add prefix to key
func (bkt *SyncBucket) Put(key string) {
	bkt.Lock()
	defer bkt.Unlock()
	bkt.Set.Put(bkt.Name + key)
}

// Keys return all keys
func (bkt *SyncBucket) Keys() (result []string) {
	bkt.RWMutex.RLock()
	defer bkt.RWMutex.RUnlock()
	lenName := len(bkt.Name)
	for _, p := range bkt.Set.pages {
		for i, key := range p.items {
			if i >= p.numItems {
				break
			}
			if strings.HasPrefix(key, bkt.Name) {
				result = append(result, key[lenName:])
			}
		}
	}
	return result
}

// Last find last key with bucket prefix
func (bkt *SyncBucket) last() (result string, idxPage, idxItem int) {
	bkt.RLock()
	defer bkt.RUnlock()

	set := bkt.Set
	key := bkt.Name
	i := sort.Search(len(set.idxs), func(n int) bool {
		if len(set.idxs[n]) > len(key) {
			return set.idxs[n][:len(key)] <= key
		}
		return set.idxs[n] <= key
	})
	if i < len(set.idxs) && set.idxs[i] == key {
		// key is present at data[i], nothing to do here
		idxPage = i
	} else {
		idxPage = i / 2
		if i == len(set.idxs) {
			//not found - append to last
			idxPage = len(set.pages) - 1
		}
	}

	//Page
	p := set.pages[idxPage]
	i = sort.Search(p.numItems, func(n int) bool {
		if len(p.items[n]) > len(key) {
			return p.items[n][:len(key)] <= key
		}
		return p.items[n] <= key
	})
	if i < p.numItems && p.items[i] == key {
		// key is present at data[i], nothing to do here
		idxItem = i
	}
	if i == p.numItems {
		// not found, new min, append at the end
		println("fn", i)
		idxItem = p.numItems
	}
	//insert or prepend

	idxItem = i
	result = ""
	if idxItem < p.numItems {
		result = p.items[idxItem]
	}

	return result, idxPage, idxItem
}

// Cursor creates a cursor associated with the bucket.
func (bkt *SyncBucket) Cursor() *Cursor {
	// Allocate and return a cursor.
	return &Cursor{
		bucket: bkt,
	}
}

// Last moves the cursor to the last item  and returns its key.
func (c *Cursor) Last() (key string) {
	result, idxPage, idxItem := c.bucket.last()
	if !strings.HasPrefix(result, c.bucket.Name) {
		return ""
	}
	c.bucket.idxPage = idxPage
	c.bucket.idxItem = idxItem
	return result[len(c.bucket.Name):]
}

// Prev moves the cursor to the previous item and returns its key.
func (bkt *SyncBucket) Prev() (key string) {
	bkt.RLock()
	defer bkt.RUnlock()
	p := bkt.Set.pages[bkt.idxPage]
	if p == nil {
		return ""
	}
	if bkt.idxItem < p.numItems-1 {
		bkt.idxItem++
		if !strings.HasPrefix(p.items[bkt.idxItem], bkt.Name) {
			return ""
		}
		return p.items[bkt.idxItem][len(bkt.Name):]
	}
	if bkt.idxPage < len(bkt.Set.pages)-1 {
		bkt.idxPage++
		bkt.idxItem = 0
		result := bkt.Set.pages[bkt.idxPage].items[bkt.idxItem]
		if !strings.HasPrefix(result, bkt.Name) {
			return ""
		}
		return result[len(bkt.Name):]
	}
	return ""
}

// Prev moves the cursor to the previous item and returns its key.
func (c *Cursor) Prev() (key string) {
	return c.bucket.Prev()
}

// ast moves the cursor to the last item  and returns its key.
func (c *Cursor) seek() (key string) {

	return
}
