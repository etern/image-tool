package main

import (
	"testing"
)

func TestNormalizeUrl(t *testing.T) {
	parent_url := "https://www.golang.org/doc.html"
	cases := []struct {
		in, out string
	}{
		{"/relative/sub.php", "https://www.golang.org/relative/sub.php"},
		{"/relative.html", "https://www.golang.org/relative.html"},
	}
	for _, c := range cases {
		got := NormalizeUrl(c.in, parent_url)
		if got != c.out {
			t.Errorf("NormalizeUrl(%q)==%q, want %q", c.in, got, c.out)
		}
	}
}

func TestFindImages(t *testing.T) {
	_, err := FindImages("http://pic.baidu.com")
	if err != nil {
		t.Error(err)
	}
}

func TestFindLinks(t *testing.T) {
	_, err := FindLinks("http://pic.baidu.com")
	if err != nil {
		t.Error(err)
	}
}
