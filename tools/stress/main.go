/*
Simple tool to simply throw requests at wysteria by the ton.

We don't care to check for any errors here .. so SIGSEGVs can (and sometimes do) occur. Meh.
*/

package main

import (
	"fmt"
	wysteria "github.com/voidshard/wysteria/client"
	"log"
	"math/rand"
	"sync"
)

func spam(i, j int) {
	client, err := wysteria.New()
	if err != nil {
		panic(err) // one err we will check -- we kinda need the client
	}

	name := fmt.Sprintf("spam_%d_%d_%d", i, j, rand.Int())
	col, _ := client.CreateCollection(name)
	log.Println("[create] collection")

	var last_item *wysteria.Item
	var last_ver *wysteria.Version

	for k := 0; k < rand.Intn(15); k++ {
		si := fmt.Sprintf("%d", i)
		sj := fmt.Sprintf("%d", j)
		sk := fmt.Sprintf("%d", k)

		m := map[string]string{
			"i": si,
			"j": sj,
			"k": sk,
		}

		item, _ := col.CreateItem(
			fmt.Sprintf("item_%d_%d_%d", i, j, k),
			fmt.Sprintf("var_%d_%d_%d", i, j, k),
			m,
		)
		log.Println("[create] item")

		if last_item != nil {
			last_item.LinkTo(fmt.Sprintf("%d%d%d", i, j, k), item)
			log.Println("[create] link")
		}
		last_item = item

		for x := 0; x < rand.Intn(20); x++ {
			client.Search(wysteria.HasFacets(m)).FindItems(
				wysteria.Limit(1+rand.Intn(10)*rand.Intn(10)),
				wysteria.Offset(rand.Intn(10)),
			)
			log.Println("[search] item")
		}

		for l := 0; l < rand.Intn(10); l++ {
			ver, _ := item.CreateVersion(m)
			log.Println("[create] version")
			for p := 0; p < rand.Intn(5); p++ {
				ver.AddResource(si, sj, sk)
				log.Println("[create] resource")
			}
			ver.Publish()
			log.Println("[publish] version")

			if last_ver != nil {
				last_ver.LinkTo(fmt.Sprintf("%d%d%d", i, j, k), last_ver)
				log.Println("[create] link")
			}
			last_ver = ver

			ver.Resources()
			log.Println("[find] resource")
			ver.Linked()
			log.Println("[find] link")
		}

		last_item.PublishedVersion()
		log.Println("[published] version")

		for y := 0; y < rand.Intn(20); y++ {
			client.Search(wysteria.HasFacets(m)).FindVersions(
				wysteria.Limit(1+rand.Intn(10)*rand.Intn(10)),
				wysteria.Offset(rand.Intn(10)),
			)
			log.Println("[find] version")
		}
	}
}

func main() {
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			closed_i := i
			for j := 0; j < 10000; j++ {
				spam(closed_i, j)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
