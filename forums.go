package main

import (
	"fmt"

	"github.com/anaskhan96/soup"
)

/*

Integration with the Dual Universe Forums @ https://board.dualthegame.com/

This WILL NOT post to the forums, it will only READ from the publicly available forums.

 */


type ForumIntegration struct {}


func (h *ForumIntegration) FollowUser(user string) (err error){

	h.Scrape("https://board.dualthegame.com/")

	return nil
}

func (h *ForumIntegration) Scrape(url string) (err error) {

	resp, err := soup.Get(url)
	if err != nil {
		return err
	}
	doc := soup.HTMLParse(resp)
	links := doc.Find("div", "id", "index_stats").FindAll("a")
	for _, link := range links {
		fmt.Println(link.Text(), "| Link :", link.Attrs()["href"])
	}

	return nil
}