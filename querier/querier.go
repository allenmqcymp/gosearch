package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/allenmqcymp/gosearch/indexer"
)

type rank struct {
	url   string
	score int
}

func evaluateQuery(query [][]string, index indexer.Indextype) []rank {
	ranks := make(map[string]int)
	for _, andquery := range query {
		andRank := evaluateAndQuery(andquery, index)
		for _, r := range andRank {
			if _, ok := ranks[r.url]; ok {
				ranks[r.url] += r.score
			} else {
				ranks[r.url] = r.score
			}
		}
	}

	var rankSlice []rank
	for k, v := range ranks {
		rankSlice = append(rankSlice, rank{k, v})
	}

	return rankSlice
}

// takes in an AND query, which is a slice of individual words related by AND
// returns a slice of tuples, each tuple
// nil means that no url in the whole index matches the query string
func evaluateAndQuery(andQuery []string, index indexer.Indextype) []rank {
	exclusionDict := make(map[string]bool)
	inclusionDict := make(map[string]int)
	var ranks []rank

	// this counts the number of words to include present
	var includeWords []string

	for _, word := range andQuery {

		// if the word is a NOT
		if word[0] == '-' {
			if len(word[1:]) < 3 {
				continue
			}

			// add the url to the exclusion dict - the value doesn't matter
			// since if a url is to be excluded, then it doesn't matter how many times it's "voted" to be excluded
			// if the word doesn't appear at all then clearly nothing to exclude
			if urls, ok := index[word[1:]]; ok {
				for k := range urls {
					exclusionDict[k] = true
				}
			}
		} else {
			if len(word) < 3 {
				continue
			}

			includeWords = append(includeWords, word)
			if urls, ok := index[word]; ok {
				for url := range urls {
					if val, put := inclusionDict[url]; put {
						inclusionDict[url] = val + 1
					} else {
						inclusionDict[url] = 1
					}
				}
			} else {
				// the word doesn't exist in the index, so no urls will match, so we can return immediately
				return nil
			}
		}
	}

	// get the urls that can be ranked
	// check that the url has all the includeWords and is not in exclusionDict
	for k, v := range inclusionDict {
		if v == len(includeWords) && !exclusionDict[k] {
			ranks = append(ranks, rank{k, math.MaxInt64})
		}
		if v > len(includeWords) {
			panic(fmt.Errorf("observed value: %d is greater than number of include words %d", v, len(includeWords)))
		}
	}

	// now we can rank the urls
	for i := range ranks {
		for _, w := range includeWords {
			score := index[w][ranks[i].url]
			ranks[i].score = min(ranks[i].score, score)
		}
	}

	return ranks
}

// quick min function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseQuery(line string) ([][]string, bool) {

	// first, split by ORs

	// we note that an or is always surrounded by two spaces
	andQueries := strings.Split(line, " or ")

	temp := strings.Fields(line)
	orCount := 0
	for _, w := range temp {
		if w == "or" {
			orCount++
		}
	}

	// usually, should be 1 more orSegments than orCount
	if orCount+1 != len(andQueries) {
		return nil, false
	}

	// then, split each AND query

	queries := make([][]string, len(andQueries))
	for i, andQ := range andQueries {
		// check each and query is well-formed
		var andQuery []string
		andCount := 0
		// now split by whitespace - since whitespace is counted as an AND implicitly
		temp := strings.Fields(andQ)
		for _, elem := range temp {
			if elem != "and" {
				andQuery = append(andQuery, elem)
			} else {
				andCount++
			}
		}
		if andCount+1 != len(strings.Split(andQ, " and ")) {
			return nil, false
		}
		queries[i] = andQuery
	}

	return queries, true
}

// takes in a query string

func containsIllegal(line string) bool {
	// if the word contains non alphabetical characters, then ignore it
	reg, err := regexp.Compile("[^a-zA-Z[:space:]-]+")
	if err != nil {
		return true
	}

	pw := reg.FindAllIndex([]byte(line), -1)
	if pw != nil {
		return true
	}

	return false
}

func main() {
	flag.Usage = func() {
		fmt.Printf("---------- Usage ---------\n")
		fmt.Printf("%s [index.json] \n", os.Args[0])
		fmt.Printf("subsequently: > [query ::= <andsequence> [or <andsequence>]\n")
		fmt.Printf("andsequence ::= <word> [and <andsequence>]\n")
		fmt.Printf("---------- Rules ---------\n")
		fmt.Printf("- * queries are converted to lowercase\n")
		fmt.Printf("- * reserved words and, - (not), or\n")
		fmt.Printf("- * if no operator is included in between search words, implicit 'and' is assumed\n")
		fmt.Printf("- * precedence from highest to lowest: - (not), and, or\n")
		fmt.Printf("- * words with less than 3 characters are ignored\n")
		fmt.Printf("---------- Examples ---------\n")
		fmt.Printf("query string: NOT word1 and word2 OR word 3 == ((not word1) and word2) or word3\n")
		fmt.Printf("query string: not this that == (not this) and that\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	indexpath := flag.Arg(0)

	// check if indexname can be loaded into an index
	// check if the fname ends with json
	if !strings.HasSuffix(indexpath, "json") {
		fmt.Printf("%s does not end with json\n", indexpath)
		os.Exit(1)
	}

	// try and load the index up
	index, err := indexer.IndexLoad(indexpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Successfully loaded index; number of unique words ", len(index))

	scanner := bufio.NewScanner(os.Stdin)
	// count the number of unique words in index
	fmt.Printf("> ")
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ToLower(line)
		if len(line) < 0 {
			fmt.Printf("> ")
			continue
		}
		if containsIllegal(line) {
			fmt.Println("invalid query (illegal characters. use only whitespace, alphabet, and dash)")
			fmt.Printf("> ")
			continue
		}
		queries, ok := parseQuery(line)
		if !ok {
			fmt.Println("invalid query (invalid syntax)")
			fmt.Printf("> ")
			continue
		}
		ranks := evaluateQuery(queries, index)
		// sort by rank
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].score > ranks[j].score
		})
		fmt.Println("---- Search results -------")
		if ranks == nil {
			fmt.Printf("no results found :(\n")
		} else {
			for _, r := range ranks {
				fmt.Printf("url: %s, rank: %d\n", r.url, r.score)
			}
		}
		fmt.Printf("> ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
