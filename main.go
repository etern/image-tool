package main

import (
	"fmt"
	"flag"
	"log"
	"path"
	"strings"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/base64"
	"github.com/PuerkitoBio/goquery"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func normalizeUrl(s string, parent_url string) string {
	u, _ := url.Parse(s)
	if !u.IsAbs() { // must before assign Scheme
		pu, _ := url.Parse(parent_url)
		u.Host = pu.Host
	}
	if len(u.Scheme) == 0 {
		u.Scheme = "http"
	}
	return u.String()
}

func findImages(url string) (images []string, err error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Println("findImages", err)
		return nil, err
	}

	doc.Find("html body img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		src = strings.Trim(src, " ")
		if len(src) == 0 || strings.HasPrefix(src, "data:image") {
			return // neglect base64 embeded image
		}
		images = append(images, normalizeUrl(src, url))
	})
	return images, nil
}

func findLinks(url string) (links []string, err error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Println("findLinks", err)
		return nil, err
	}

	doc.Find("html body a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		href = strings.Trim(href, " ")
		if len(href) == 0 || strings.HasPrefix(href, "javascript:") {
			return
		}
		links = append(links, normalizeUrl(href, url))
	})
	return links, nil
}

func getImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("http Get", url, err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println("status code error:", resp.StatusCode, resp.Status)
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

type WebImageFs struct {
	pathfs.FileSystem
	rootSite string
	images map[string]string
}

func (me *WebImageFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	name = path.Base(name)
	if url, ok := me.images[name]; ok {
		data, err := getImage(url) // TODO: caching
		if err != nil {
			return nil, fuse.ENOENT
		}
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644, Size: uint64(len(data)),
		}, fuse.OK
	} else {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (me *WebImageFs) OpenDir(name string, context *fuse.Context) (entries []fuse.DirEntry, code fuse.Status) {
	url := name
	if name == "" {
		url = me.rootSite
	} else {
		url_raw, _ := base64.StdEncoding.DecodeString(name)
		url = string(url_raw) // utf-8
	}
	images, err := findImages(url)
	if err != nil {
		return nil, fuse.ENOENT
	}
	me.images = make(map[string]string)
	for _, img := range images {
		nm := path.Base(img)
		me.images[nm] = img // TODO: check duplication
		entries = append(entries, fuse.DirEntry{Name: nm, Mode: fuse.S_IFREG})
	}

	links, err := findLinks(url)
	if err != nil {
		return nil, fuse.ENOENT
	}
	for _, link := range links {
		nm := base64.StdEncoding.EncodeToString([]byte(link)) // utf-8 bytes
		entries = append(entries, fuse.DirEntry{Name: nm, Mode: fuse.S_IFDIR})
	}
	return entries, fuse.OK
}

func (me *WebImageFs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	name = path.Base(name)
	url := me.images[name]
	data, err := getImage(url)
	if err != nil {
		return nil, fuse.ENOENT
	}
	return nodefs.NewDataFile(data), fuse.OK
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 2 {
		log.Fatal("Usage:\n  image_tool <mountpoint> <website>")
	}
	webImgFs := WebImageFs{
		FileSystem: pathfs.NewReadonlyFileSystem(pathfs.NewDefaultFileSystem()),
		rootSite: flag.Arg(1),
	}
	nfs := pathfs.NewPathNodeFs(&webImgFs, nil)
	server, _, err := nodefs.MountRoot(flag.Arg(0), nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	server.Serve()
}
