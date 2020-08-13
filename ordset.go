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
	sync.RWMutex
	idxs  []string
	pages []*page
}

type SyncBucket struct {
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
func (set *OrdSet) Put(key string) {
	set.Lock()
	defer set.Unlock()
	set.put(key)
}

// Put will add key in set, if not present
func (set *OrdSet) put(key string) {
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
	if set.pages[idx].numItems == pageSize-1 {
		set.split(idx)
		set.put(key)
		return
	}
	set.pages[idx].add(key)
}

func (p *page) add(key string) *page {
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
		//fmt.Println("data i == p.numItems:", p.items, p.min, p.max, p.numItems)
		return p
	}

	//insert or prepend
	if i == 0 {
		//prepend, new max
		p.max = key
		p.min = p.items[p.numItems-1]
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
	for i := mid; i < pageSize; i++ {
		p.items[i] = ""
	}
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
	set.RLock()
	defer set.RUnlock()
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
	bkt.Set.Put(bkt.Name + key)
}

//1597323014317877000
//&{items:[userrob userpike userbob useranna useralice user01 itemzxlxibiq itemzwqwagcl itemzubqeabj itemztejlftb itemzoigtdql itemzmbixqpj itemzlzafkqq itemzefhtvej itemzbblspnm itemyscgyhze itemymxptnbi itemycthwabh itemxxpiajhh itemxwmqvlay itemxwmnciuo itemxvhhdwtu itemxvfgxlfe itemxlobjqak itemxhszckaq itemxfmxdowt itemxdvsxydz itemxdbvmmsv itemxcazsohq itemwzydxbug itemwxywyikw itemwwnqpprh itemwqmlaszl itemwmlnsxls itemwjmkcypk itemwielsnjv itemwckthyww itemvpvatlbl itemvogmqzvy itemvnwvcaep itemvlwicokh itemvggkjawz itemvgeprumq itemvfxjlxap itemvfwohngq itemvblmjczj itemuyqzncjk itemuulksexn itemunjqbjdi itemulzbnmzb itemuixczudg itemudrsylfz itemubxerokd itemttppgnlw itemtsggxmih itemtoshfqel itemtlmrjgri itemtjsqynxf itemthpvmwtb itemtggplmvm itemtfhblawi itemteihpnwd itemtcijmude itemtajvccsq itemszbltmfi itemswlvtgrw itemstaqmian itemsrihktjd itemspvuggla itemspopzdet itemspgtokbb itemskytmqkl itemsfatlllk itemseoujhko itemseefcsnu itemsdmrvxve itemsabyckur itemrxsyeglq itemrtpldeog itemrqwymfsf itemroqqegct itemrneuraro itemrlrtkwoh itemrknzubru itemrjqbbuvy itemrilncuqo itemrerxvpyf itemrarpscgl itemqtmtixqn itemqribxxpp itemqqgcldks itemqofipytu itemqmyrnbxv itemqlvcroth itemqfwmgbdw itempwlqsfnn itempslslbnp itemppdkblzt itempnyzsyzd itempnapltiv itemplvyqziy itempkzqkmwd itempjxtyyrh itempiqhahsn itempimkvewn itemphtvsxkz itemphmspsgo itempfaarruq itempefmijla itempecqafzz itemoyymwymc itemoweshgus itemovryokod itemovilxocr itemospiruhr itemopxscsds itemonzqoont itemomjnbdcv itemolbbjiij itemoiackjea itemogasoxmg itemobilbfrl itemoavbkcgy itemnzkcleve itemnwqiricf itemnuvkifap itemnthhjmqp itemnouthzto itemnoizbcsw itemnkpzajds itemnkafkhyg itemnfaenvwa itemndregpaq itemmntsgnje itemmhaogmeh itemmanrliop itemmaijkutz itemlrmevskn itemlretoznv itemlmiporya itemlmimkzgj itemllhhuopr itemljnpkonq itemlixyrbcd itemlfhtyhym itemldbnscay itemlanrjacx itemkwthrpui itemkvjtarff itemkumfrzss itemkrzkkkfi itemkqvgzudj itemkoliqwbs itemkmrgbxta itemkkniepcx itemkeqyimfb itemjvblgazc itemjtlbpftp itemjhavnsrm itemjfesbyzh itemjdcyxtfr itemjbxbydof itemiupbyqeo itemirzkmzpk itemirfroekw itemiqvuszwh itemiqcyszkn iteminnskdgs itemijvbiscv itemiiiketzy itemihocnhsj itemifqszbss itemibclkods itemhxihpszv itemhwedbmqb itemhumycbrn itemhpjvkkgm itemhmgninia itemhkjgqknt itemhibvlriu itemhhwhawsm itemhgrzyemt itemhfdvqarg itemhejdvwik itemhegvmudm itemhcjvabdw itemhbcnnjmv itemgzzzrdae itemgyhnpooc itemgxykqhuw itemgxkvqzbt itemgubqvlts itemgrudpotl itemgnaiwfcj itemgmstuefj itemgmbgwevc itemgjwhdivt itemgixdsrox itemghobzvzn itemgefceqbo itemgecfklmi itemgdgkfvao itemgcntyaam itemfzekuiek itemfsnigwxb itemfmsapaqx itemfmlxfesn itemflyyldte itemfihwvajb itemfibjxolh itemfgszfenf itemfbfmfwwn itemfarfjckz itemexrislmc itemevugsoko itememzbxsvv itememetlilk itemehtixtpi itemecxvrobb itemecewcqfm itemdzcyzcta itemdtbgqdwh itemdrpsvicf itemdptzjtmg itemdpeefkmc itemdmoqluuu itemdjtkdzgu itemdjqbmkcl itemdjoockbl itemdbfrqhkk itemcrviviay itemcqnlscek itemckfdzdpz itemcdqhydng itemcctghqgt itemcbrglfxf itembudgokch itembszudsfx itembsiwxfpt itembhxgurlw itembfpmshrv itembcpykxlz itembbconwra itembaihifwa itemayjjcekr itemaxuohvic itemaxqucubc itemaxagyeds itemasofcmvl itemaqtzsgvq itemaqhhxzmg itemakhlpgns itemajzcxzrn itemafwpvxgk itemadmgfkli ] min:itemndregpaq max:userrob numItems:133 modified:false}
// Keys return all keys
func (bkt *SyncBucket) Keys() (result []string) {
	bkt.Set.RLock()
	defer bkt.Set.RUnlock()
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
	bkt.Set.RLock()
	defer bkt.Set.RUnlock()

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
		idxPage = i / 2
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
	bkt.Set.RLock()
	defer bkt.Set.RUnlock()

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

func (set *OrdSet) has(key string) bool {
	//fmt.Printf("Add %s %+v\n", key, set)
	// sort desc
	i := sort.Search(len(set.idxs), func(n int) bool {
		return set.idxs[n] <= key
	})
	if i < len(set.idxs) && set.idxs[i] == key {
		// key is present at data[i], nothing to do here
		return true
	}

	idx := i / 2
	if i == len(set.idxs) {
		//not found - append to last
		idx = len(set.pages) - 1
	}
	p := set.pages[idx]

	i = sort.Search(p.numItems, func(n int) bool {
		return p.items[n] <= key
	})
	//fmt.Println("page i", i, key, p.items[1] == key)
	if i < p.numItems && p.items[i] == key {
		// key is present at data[i], nothing to do here
		return true
	}
	return false
}

// Has return true if key in set
func (set *OrdSet) Has(key string) bool {
	set.Lock()
	defer set.Unlock()
	return set.has(key)
}

func (set *OrdSet) delete(key string) bool {
	//fmt.Printf("Add %s %+v\n", key, set)
	// sort desc
	i := sort.Search(len(set.idxs), func(n int) bool {
		return set.idxs[n] <= key
	})

	idx := i / 2
	if i == len(set.idxs) {
		//not found - append to last
		idx = len(set.pages) - 1
	}
	//p := set.pages[idx]
	i = sort.Search(set.pages[idx].numItems, func(n int) bool {
		return set.pages[idx].items[n] <= key
	})
	//fmt.Println("page i", i, key, p.items[1] == key)
	if i < set.pages[idx].numItems && set.pages[idx].items[i] == key {
		//delete
		//set.pages[idx].
		copy(set.pages[idx].items[i:], set.pages[idx].items[i+1:])
		if i == set.pages[idx].numItems-1 {
			if i == 0 {
				//delete set.pages[idx]
				fmt.Println("delete set.pages[idx]")
			} else {
				//last elem
				set.pages[idx].min = set.pages[idx].items[i-1]
				//upd index
			}
		}
		if i == 0 {
			set.pages[idx].max = set.pages[idx].items[i]
		}
		set.pages[idx].numItems--
		fmt.Printf("\n%s %+v\n", key, set.pages[idx])
		return true
	}
	return false
}

func (set *OrdSet) Delete(key string) bool {
	set.Lock()
	defer set.Unlock()
	return set.delete(key)
}
