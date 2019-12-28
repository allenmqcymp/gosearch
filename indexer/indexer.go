package indexer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/allenmqcymp/gosearch/pageio"
)

// Indextype is a composite type representing a map of words, each referring to a map of urls and frequency counts
type Indextype map[string]map[string]int

var indexm = struct {
	m   Indextype
	mux sync.Mutex
}{}

// GenerateIndex generates an index of all the encountered words and frequencies
// dir is where all the pages reside
// Returns a map of words. Each key - a word, has value of a map, which associates a url with the frequency of the word
func GenerateIndex(dir string) Indextype {

	// initialize the map
	indexm.m = make(Indextype)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan bool)
	for _, file := range files {
		fname := file.Name()
		go func() {
			err := traversePage(fname, dir)
			if err != nil {
				log.Fatal(err)
			}
			done <- true
		}()
	}
	for range files {
		<-done
	}
	return indexm.m
}

func traversePage(fname, dir string) error {
	// load up the page
	pg, err := pageio.Pageload(fname, dir)
	if err != nil {
		return err
	}

	// if the word contains non alphabetical characters, then ignore it
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		return err
	}

	words := strings.Fields(pg.Text)
	for _, w := range words {
		if len(w) <= 3 {
			continue
		}

		pw := reg.FindAllIndex([]byte(w), -1)
		if pw != nil {
			continue
		}

		// convert the word to lowercase
		w := strings.ToLower(w)

		indexm.mux.Lock()
		_, ok := indexm.m[w]
		if !ok {
			indexm.m[w] = make(map[string]int)
			indexm.m[w][pg.Url] = 1
		} else {
			indexm.m[w][pg.Url]++
		}
		indexm.mux.Unlock()
	}
	return nil
}

// IndexSave saves the index data structure into a JSON format
// GenerateIndex should be called beforehand to ensure the index is fully built
/*
	{
		word1: {
			url1: count11
			url2: count12
		}
		word2: {
			url1: count21
			url2: count22
		}
	}
*/
func IndexSave(fname string) error {

	// check if the fname ends with json
	if !strings.HasSuffix(fname, "json") {
		return fmt.Errorf("%s does not end with json", fname)
	}

	b, err := json.MarshalIndent(indexm.m, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fname, b, 0644)

	return err
}

// IndexLoad takes in the path of a index file, and returns an index, reporting any error
func IndexLoad(indexPath string) (Indextype, error) {

	// load as a bytestring
	b, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var f Indextype
	err = json.Unmarshal(b, &f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// func main() {

// 	if len(os.Args) < 3 {
// 		fmt.Println("usage: ./indexer [dirname] [fname]")
// 		return
// 	}

// 	dirname := os.Args[1]
// 	indexname := os.Args[2]

// 	// load up the index
// 	GenerateIndex(dirname)

// 	// save the index
// 	err := IndexSave(indexname)

// 	if err != nil {
// 		fmt.Println(err)
// 		fmt.Println("finished unsuccessfully")
// 	} else {
// 		fmt.Println("finished successfully!")
// 	}
// }
