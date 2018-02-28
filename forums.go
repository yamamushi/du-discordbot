package main

import (
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"strings"
)

/*

Integration with the Dual Universe Forums @ https://board.dualthegame.com/

This WILL NOT post to the forums, it will only READ from the publicly available forums.

*/

// ForumIntegration struct
type ForumIntegration struct{}

// FollowUser function
func (h *ForumIntegration) FollowUser(user string) (err error) {

	return nil
}

// Scrape function
func (h *ForumIntegration) Scrape(url string) (response string, err error) {

	resp, err := soup.Get(url)
	if err != nil {
		return resp, err
	}
	/*
		doc := soup.HTMLParse(resp)
		links := doc.Find("div", "id", "index_stats").FindAll("a")
		for _, link := range links {
			fmt.Println(link.Text(), "| Link :", link.Attrs()["href"])
		}
	*/

	return resp, nil
}

// GetLatestCommentForThread function
func (h *ForumIntegration) GetLatestCommentForThread(url string) (username string, comment string, commenturl string, err error) {

	resp, err := soup.Get(url + "/&page=1000") // Append page=1000 so we get the last page
	if err != nil {
		fmt.Println("Could not retreive page: " + url)
		return "", "", "", err
	}
	doc := soup.HTMLParse(resp)
	comments := doc.Find("div", "class", "cTopic ipsClear ipsSpacer_top").FindAll("article")

	lastid := len(comments)

	if lastid > 0 {

		commentid := strings.TrimPrefix(comments[lastid-1].Attrs()["id"], "elComment_")
		latestcommentlink := url + "&do=findComment&comment=" + commentid

		usernamelement := comments[lastid-1].Find("h3", "class", "ipsType_sectionHead cAuthorPane_author ipsType_blendLinks ipsType_break").Find("a")
		username := strings.Trim(strings.TrimSuffix(strings.TrimPrefix(usernamelement.Attrs()["title"], "Go to "), "'s profile"), " ")

		comment := comments[lastid-1].Find("div", "class", "ipsType_normal ipsType_richText ipsContained").FindAll("p")

		lasttext := len(comment)
		truncatedcomment := ""
		if lasttext > 0 {
			//fmt.Println(comment.Text())
			a := []rune(comment[lasttext-1].Text())
			for i, r := range a {
				if i < 120 {
					truncatedcomment = truncatedcomment + string(r)
				}
				// every 3 i, do something
			}
			truncatedcomment = strings.TrimSuffix(strings.Trim(truncatedcomment, " "), "\n") + "..."
		}
		return username, truncatedcomment, latestcommentlink, nil
	}

	fmt.Println("Could not find comment id")
	return "", "", "", errors.New("Could not find comment id")
}
