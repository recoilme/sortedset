package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"

	"github.com/recoilme/ordset"
	"github.com/tidwall/lotsa"
)

func main() {
	ParallelOrdSet()
}

func randKeys(N int) (keys []string) {
	format := fmt.Sprintf("%%0%dd", len(fmt.Sprintf("%d", N-1)))
	for _, i := range rand.Perm(N) {
		keys = append(keys, fmt.Sprintf(format, i))
	}
	return
}

func ParallelOrdSet() {
	N := 10_000_000
	set := ordset.New()
	keys := randKeys(N)
	bkt := ordset.Bucket(set, "")
	lotsa.Output = os.Stdout
	lotsa.MemUsage = true
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		bkt.Put(keys[i])
	})

}

/*
type googleKind struct {
	key string
}

func (a *googleKind) Less(b btree.Item) bool {
	return a.key < b.(*googleKind).key
}
func ParallelGoogle() {
	N := 10_000
	fmt.Println()
	type gtree struct {
		sync.RWMutex
		gt *btree.BTree
	}
	gt := &gtree{gt: btree.New(32)}
	//tr := btree.New(32)
	keys := randKeys(N)
	gkeys := make([]*googleKind, len(keys))

	lotsa.Output = os.Stdout
	lotsa.MemUsage = true
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		gt.Lock()
		gt.gt.ReplaceOrInsert(gkeys[i])
		gt.Unlock()
	})

}
*/
