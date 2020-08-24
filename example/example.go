package main

import (
	"github.com/recoilme/sortedset"
	"golang.org/x/exp/errors/fmt"
)

func main() {
	base()
	buckets()
	cursor()
}

func base() {
	set := sortedset.New()
	set.Put("a")
	set.Put("b")
	fmt.Println(set.Keys())
	//[b a]
}

func buckets() {

	set := sortedset.New()
	users := sortedset.Bucket(set, "user")
	users.Put("rob")
	users.Put("bob")
	users.Put("pike")
	users.Put("alice")
	fmt.Println(users.Keys(0, 0))
	// output: [rob pike bob alice]
	items := sortedset.Bucket(set, "item")
	items.Put("003")
	items.Put("042")
	fmt.Println(items.Keys(0, 0))
	// output: [042 003]
}

func cursor() {
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

}
