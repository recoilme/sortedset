package main

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/recoilme/ordset"
)

var rnd *rand.Rand

func init() {
	seed := int64(int64(time.Now().UnixNano()))
	fmt.Printf("seed: %d\n", seed)
	rnd = rand.New(rand.NewSource(seed))
}

/*
func BenchmarkAddRandGoogle(b *testing.B) {
	tr := btree.New(32)
	keys := randKeys(b.N)
	gkeys := make([]*googleKind, len(keys))
	for i := 0; i < b.N; i++ {
		gkeys[i] = &googleKind{keys[i]}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.ReplaceOrInsert(gkeys[i])
	}
}*/

//go test -benchmem -bench Add
func BenchmarkAddRand(b *testing.B) {
	keys := randKeys(b.N)
	set := ordset.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}

func BenchmarkAddAsc(b *testing.B) {
	keys := randKeys(b.N)
	sort.Strings(keys)
	set := ordset.New(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}

func BenchmarkAddDesc(b *testing.B) {
	keys := randKeys(b.N)
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	set := ordset.New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Put(keys[i])
	}
}
