package pageio

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// Webpage is a struct that holds the url, page body, and depth at which it was crawled
type Webpage struct {
	Url   string
	Text  string
	Depth int
}

// Pagesave saves the url of the page, the contents of the page
// returns any error if unsuccessful
func Pagesave(webpage *Webpage, name string, dir string) error {
	if e, _ := exists(dir); !e {
		return fmt.Errorf("dir %s does not exist", dir)
	}
	fpath := path.Join(dir, name)
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer f.Close()
	line := []byte(webpage.Url + "\n")
	_, err = f.Write(line)
	if err != nil {
		return err
	}
	line = []byte(strconv.Itoa(webpage.Depth) + "\n")
	_, err = f.Write(line)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(webpage.Text))
	if err != nil {
		return err
	}
	return nil
}

// Pageload takes in dir/name of the page, and returns a pointer to populated webpage struct, and err, if any
func Pageload(name string, dir string) (*Webpage, error) {
	if e, _ := exists(dir); !e {
		return nil, fmt.Errorf("dir %s does not exist", dir)
	}
	fpath := path.Join(dir, name)
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	res := scanner.Scan()
	if !res {
		return nil, scanner.Err()
	}
	url := scanner.Text()

	res = scanner.Scan()
	if !res {
		return nil, scanner.Err()
	}
	depth, err := strconv.Atoi(scanner.Text())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// read the rest of the body into
	var body strings.Builder
	for scanner.Scan() {
		fmt.Fprintf(&body, "%s\n", scanner.Text())
	}

	// since we artificially add \n, we have to remove the last \n, since the last segment resulted
	// from an EOF and not a newline
	strText := body.String()[:len(body.String())-1]
	return &Webpage{Url: url, Text: strText, Depth: depth}, nil

}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
