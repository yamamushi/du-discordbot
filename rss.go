package main

import (
	"github.com/mmcdole/gofeed"
	"fmt"
)

type RSS struct {

	Feeds RSSFeeds
}


type RSSFeed struct {
	Name 	string
	URL   string
}


type RSSFeeds struct {
	Feeds[] RSSFeed
}


func (h *RSS) Validate (url string) error {

	fmt.Println(url)
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return err
	}


	fmt.Println(feed.Title)
	fmt.Println(feed.Link)

	return nil

}