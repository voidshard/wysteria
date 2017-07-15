/*
Simple tool to simply throw requests at wysteria by the ton.

We don't care to check for any errors here .. so SIGSEGVs can (and sometimes do) occur. Meh.
*/

package main

import (
	wysteria "github.com/voidshard/wysteria/client"
	"sync"
	"fmt"
	"math/rand"
)

func spam(i, j int) {
	client, err := wysteria.New()
	if err != nil {
		panic(err) // one err we will check -- we kinda need the client
	}

	name := fmt.Sprintf("spam_%d_%d_%d", i, j, rand.Int())
	col, _ := client.CreateCollection(name)

	var last_item *wysteria.Item
	var last_ver *wysteria.Version

	for k := 0 ; k < 10; k ++ {
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

		if last_item != nil {
			last_item.LinkTo(fmt.Sprintf("%d%d%d", i, j, k), item)
		}
		last_item = item

		for l := 0 ; l < 10; l ++ {
			ver, _ := item.CreateVersion(m)
			for p := 0 ; p < 5; p ++ {
				ver.AddResource(si, sj, sk)
			}

			if last_ver != nil {
				last_ver.LinkTo(fmt.Sprintf("%d%d%d", i, j, k), last_ver)
			}
			last_ver = ver
		}
	}
}

func main () {
	wg := sync.WaitGroup{}

	for i := 0 ; i < 10; i ++ {
		wg.Add(1)
		go func() {
			closed_i := i
			for j := 0 ; j < 10; j ++ {
				spam(closed_i, j)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}