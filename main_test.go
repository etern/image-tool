package main

import (
	"fmt"
	"testing"
)

func TestGetImages(t *testing.T) {
	images, err := GetImages("http://pic.baidu.com")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(images)
}

func TestGetLinks(t *testing.T) {
	links, err := GetLinks("http://pic.baidu.com")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(links)
}
