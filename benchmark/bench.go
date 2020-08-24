package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"

	"github.com/google/btree"
	"github.com/recoilme/ordset"
	"github.com/tidwall/lotsa"
)

func main() {

	OneThreadOrdSet()
	OneThreadGoogle()
	ParallelOrdSet()
}

func randKeys(N int) (keys []string) {
	format := fmt.Sprintf("%%0%dd", len(fmt.Sprintf("%d", N-1)))
	for _, i := range rand.Perm(N) {
		keys = append(keys, fmt.Sprintf(format, i))
	}
	return
}

func OneThreadOrdSet() {
	fmt.Println("OneThreadOrdSet")
	N := 10_000_000
	set := ordset.New()
	keys := randKeys(N)
	bkt := ordset.Bucket(set, "")
	lotsa.Output = os.Stdout
	lotsa.MemUsage = true
	lotsa.Ops(N, 1, func(i, _ int) {
		bkt.Put(keys[i])
	})
}

func ParallelOrdSet() {
	fmt.Println("ParallelOrdSet")
	N := 10_000_000
	set := ordset.New()
	keys := randKeys(N)
	lotsa.Output = os.Stdout
	lotsa.MemUsage = true
	lotsa.Ops(N, runtime.NumCPU(), func(i, _ int) {
		set.Put(keys[i])
	})
}

// Str implements the Item interface for strings.
type Str string

// Less returns true if a < b.
func (a Str) Less(b btree.Item) bool {
	return a < b.(Str)
}

var btreeDegree = flag.Int("degree", 32, "B-Tree degree")

func OneThreadGoogle() {
	fmt.Println("OneThreadGoogle")
	N := 10_000_000
	keys := randKeys(N)
	tr := btree.New(*btreeDegree)
	lotsa.Output = os.Stdout
	lotsa.MemUsage = true
	lotsa.Ops(N, 1, func(i, _ int) {
		tr.ReplaceOrInsert(Str(keys[i]))
	})
}
