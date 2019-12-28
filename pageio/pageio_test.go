package pageio_test

import (
	"testing"

	"github.com/allenmqcymp/gosearch/pageio"
)

var dummyWebpage = pageio.Webpage{
	Url: "https://dummywebsite.com",
	Text: `yp��w-��]�	A��®�	$�b�:c(�	�R�1ԓ0C
		
		
		`,
	Depth: 2,
}

func Test(t *testing.T) {

	err := pageio.Pagesave(&dummyWebpage, "testpg", ".")
	if err != nil {
		t.Error("failed to save page\n")
	}

	pg, err := pageio.Pageload("testpg", ".")
	var expectations = []struct {
		s, want string
	}{
		{pg.Url, dummyWebpage.Url},
		{pg.Text, dummyWebpage.Text},
	}
	for _, c := range expectations {
		if c.s != c.want {
			t.Errorf("pageloaded %q != original %q", c.s, c.want)
		}
	}
}
