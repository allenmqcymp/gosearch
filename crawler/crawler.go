package crawler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/allenmqcymp/gosearch/pageio"
)

// global variables
var fetched = struct {
	m   map[string]string
	mux sync.Mutex
	id  int
}{m: make(map[string]string), id: 0}

var seedURL = "https://thayer.github.io/engs50/"
var fileDir = "../pages/"
var loading = "loading"

// Fetcher returns the body of URL and
// a slice of URLs found on that page, any errors associated as well
type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {

	if depth < 0 {
		return
	}

	sz := len(url)
	expurl := url + string('/')
	stripurl := url[:sz-1]

	fetched.mux.Lock()
	if _, ok := fetched.m[url]; ok {
		fetched.mux.Unlock()
		return
	}
	if _, ok := fetched.m[expurl]; ok {
		fetched.mux.Unlock()
		return
	}
	if _, ok := fetched.m[stripurl]; ok {
		fetched.mux.Unlock()
		return
	}
	// we give the url a loading status to prevent other goroutines from loading it
	fetched.m[url] = loading
	fetched.mux.Unlock()

	body, urls, err := fetcher.Fetch(url)

	fetched.mux.Lock()
	if err != nil {
		fmt.Printf("failed to load %s\n", url)
		// if the load is unsuccessful, then we delete the loading entry
		delete(fetched.m, url)
		fetched.mux.Unlock()
		return
	}
	fetched.m[url] = url
	curid := fetched.id
	fetched.id++
	fetched.mux.Unlock()

	// now we save this newpage
	newPage := pageio.Webpage{Url: url, Text: body, Depth: depth}
	err = pageio.Pagesave(&newPage, strconv.Itoa(curid), fileDir)
	fmt.Printf("first time done with %s at depth %d\n", url, depth)

	if err != nil {
		log.Println(err)
		log.Printf("failed to save %s\n", url)
	}

	var channel = make(chan string)
	for _, u := range urls {
		go func(u string) {
			Crawl(u, depth-1, fetcher)
			channel <- u
		}(u)
	}
	for i := range urls {
		fmt.Printf("[%v] Waiting for child %v\n", url, i)
		<-channel
	}
}

// PFetcher is a type of fetcher that fetches urls
// in this case PFetcher is merely used as a placeholder
type PFetcher struct {
}

// isInternalURL checks if the url is a subdomain of seedURL
func isInternalURL(url string) bool {
	return strings.HasPrefix(url, seedURL)
}

// Fetch takes in a url, fetches the body of the page associated with the url
// as well as returns all other urls on the page
// and any error associated with fetching
func (p PFetcher) Fetch(url string) (string, []string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("http.Get() failed with '%s'\n", err)
		return "", nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			return "", nil, err
		}
		bodyString := string(bodyBytes)
		var urls []string
		n := 0
		done := false
		nextURL := ""
		for {
			// fetch the next url
			nextURL, n, done = fetchNextURL(bodyString, n)
			if done {
				break
			}
			if okURL, ok := okNormalizeNextURL(nextURL, url); ok {
				urls = append(urls, okURL)
			}
		}
		return bodyString, urls, nil
	}
	return "", nil, fmt.Errorf("resp status code not OK %d", resp.StatusCode)
}

// checks if the url is OK and NORMALIZES it
// takes in the url to be analyzed, cururl, as well as the baseurl (the URL of the page from which curlurl was found)
// checks if cururl is a legitimate new url, ie.
// URLS that it ignores are ones that are RELATIVE and that start with a # - since those simply point to different areas on the same page
// Resolves relative urls into absolute urls
// and removes the trailing slash from urls
func okNormalizeNextURL(cururl string, baseurl string) (string, bool) {

	if len(cururl) == 0 {
		return "", false
	}

	if string(cururl[0]) == "#" {
		return "", false
	}

	baseURLStruct, err := url.Parse(baseurl)
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	urlStruct, err := url.Parse(cururl)
	if err != nil {
		fmt.Println(err)
		return "", false
	}
	newURLStruct := baseURLStruct.ResolveReference(urlStruct)

	if !isInternalURL(newURLStruct.String()) {
		return "", false
	}

	// add a trailing slash - because sometimes links like labs/lab2/ and labs/lab2 are counted twice
	// whereas removing the slash sometimes doesn't count the url as a valid directory url at all
	// this is too complex to deal with ad-hoc - read RFC manual if interested - I don't have that much time to dive into it
	retURL := newURLStruct.String()

	// exclude fragments, portions of the url that begin with #
	fragid := strings.Index(retURL, "#")

	if fragid != -1 {
		retURL = retURL[:fragid]
	}

	return retURL, true
}

// fetchNextURL takes in HTML and an integer specifying the start of the HTML to search (start is inclusive)
// Returns the next url found, parsing from left to right, top to bottom
// Returns true when done, otherwise false, the next found url, and the index where the next found url ends
func fetchNextURL(html string, idx int) (string, int, bool) {
	body := html[idx:]
	startLink := strings.Index(body, "a href")
	if startLink == -1 {
		return "", 0, true
	}
	// basically, we want to find the blah.com in a 'a href="blah.com"'
	startIndex := strings.Index(body[startLink:], "\"")
	startIndex += startLink

	endIndex := strings.Index(body[startIndex+1:], "\"")
	endIndex += startIndex + 1

	newURL := body[startIndex+1 : endIndex]

	return newURL, idx + endIndex, false
}

// func main() {

// 	if len(os.Args) < 2 {
// 		fmt.Println("usage: ./crawler [maxdepth] [seedURL (default: https://thayer.github.io/engs50/)]")
// 		return
// 	}

// 	maxdepth := 0
// 	var err error

// 	if len(os.Args) == 2 {
// 		maxdepth, err = strconv.Atoi(os.Args[1])
// 	} else {
// 		maxdepth, err = strconv.Atoi(os.Args[1])
// 		seedURL = os.Args[2]
// 	}

// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	fetcher := PFetcher{}
// 	// url, maxdepth, call the fetcher function to use
// 	Crawl(seedURL, maxdepth, fetcher)

// 	// print stats
// 	fmt.Println("Fetching stats\n--------------")
// 	for url := range fetched.m {
// 		fmt.Printf("%v was fetched\n", url)
// 	}
// 	fmt.Printf("total number fetched %d\n", len(fetched.m))
// }
