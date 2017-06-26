package main

import (
	"github.com/yamamushi/gofeed"
	"time"
	"errors"
	"fmt"
	"strings"
)

type RSS struct {
	db		*DBHandler
}


type RSSFeed struct {
	ID	string `storm:"id"`
	Title 	string
	URL   string `storm:"index"`
	LastRun time.Time // Not particularly needed, but may become useful
	Created string
	Description string
	Author string
	Updated string
	Published string
	ChannelID string `storm:"index"`// Limit our feeds per channel
	Link string
	Twitter bool `storm:"index"`
	Reddit bool
	LastItem string // URL of last item
	Posts []string // List of Posted URL's
}

type RSSItem struct {

	Title string
	Author string
	Content string
	Description string
	Published string
	Link string
	Twitter bool
	Reddit bool

}


func (h *RSS) Validate (url string) error {

	fp := gofeed.NewParser()
	_, err := fp.ParseURL(url)
	if err != nil {
		return err
	}

	return nil
}


func (h *RSS) AddToDB(rssfeed RSSFeed) (err error){
	db := h.db.rawdb.From("RSS")
	err = db.Save(&rssfeed)
	return err
}


func (h *RSS) GetFromDB(url string, channel string) (rssfeed RSSFeed, err error){

	db := h.db.rawdb.From("RSS")
	rssfeeds := []RSSFeed{}
	err = db.Find("URL", url, &rssfeeds)
	if err != nil {
		return rssfeed, err
	}
	if len(rssfeeds) < 1 {
		return rssfeed, errors.New("No record exists")
	}

	for _, i := range rssfeeds {
		if i.ChannelID == channel {
			return i, nil
		}
	}

	return rssfeed, errors.New("No record found")
}

func (h *RSS) GetDB() (rssfeeds []RSSFeed, err error){
	db := h.db.rawdb.From("RSS")
	err = db.All(&rssfeeds)
	if err != nil{
		return rssfeeds, err
	}
	return rssfeeds, nil
}

func (h *RSS) GetChannel(channel string) (rssfeeds []RSSFeed, err error){

	db := h.db.rawdb.From("RSS")
	err = db.Find("ChannelID", channel, &rssfeeds)
	if err != nil {
		return rssfeeds, err
	}
	if len(rssfeeds) < 1 {
		return rssfeeds, errors.New("No record exists")
	}

	return rssfeeds, nil
}


func (h *RSS) CheckDB (url string, channel string) bool {

	_, err := h.GetFromDB(url, channel)
	if err != nil {
		return true
	}

	return false
}



func (h *RSS) GetTitle(url string, channel string) (title string, err error){

	rssfeed, err := h.GetFromDB(url, channel)
	if err != nil {
		if rssfeed.URL == url {
			return rssfeed.Title, nil
		}
	} // If we don't error, that could mean we don't have it in the database yet, so keep going...

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return "", err
	}

	title = feed.Title
	return title, nil

}

func (h *RSS) GetDescription(url string, channel string) (description string, err error){

	rssfeed, err := h.GetFromDB(url, channel)
	if err != nil {
		 if rssfeed.URL == url {
			 return rssfeed.Description, nil
		 }
	} // If we don't error, that could mean we don't have it in the database yet, so keep going...

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return "", err
	}

	description = feed.Description
	return description, nil

}

func (h *RSS) GetUpdated(url string, channel string) (updated string, err error){

	rssfeed, err := h.GetFromDB(url, channel)
	if err != nil {
		if rssfeed.URL == url {
			return rssfeed.Updated, nil
		}
	} // If we don't error, that could mean we don't have it in the database yet, so keep going...

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return "", err
	}

	updated = feed.Updated
	return updated, nil


}

func (h *RSS) GetPublished(url string, channel string) (published string, err error){

	rssfeed, err := h.GetFromDB(url, channel)
	if err != nil {
		if rssfeed.URL == url {
			return rssfeed.Published, nil
		}
	} // If we don't error, that could mean we don't have it in the database yet, so keep going...

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return "", err
	}

	published = feed.Published
	return published, nil

}

/*
func (h *RSS) GetAuthor(url string, channel string) (author string, err error){

	rssfeed, err := h.GetFromDB(url, channel)
	if err != nil {
		if rssfeed.URL == url {
			return rssfeed.Author, nil
		}
	} // If we don't error, that could mean we don't have it in the database yet, so keep going...

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return "", err
	}

	author = feed.Author.Name
	return author, nil
}
*/

func (h *RSS) Subscribe(url string, title string, channel string) (err error){

	if !h.CheckDB(url, channel){
		return errors.New("Already Subscribed to: " + url)
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseUserAgentURL(url, "Dual_Universe_Discord_Bot/1.0")
	if err != nil {
		return err
	}

	if title == "" {
		title = feed.Title
	}

	var author string
	if feed.Author != nil {
		author = feed.Author.Name
	} else {
		author = ""
	}

	isTwitter := false
	if strings.HasPrefix( url, "https://twitrss.me/") {
		author = strings.TrimPrefix(feed.Link, "https://twitter.com/")
		isTwitter = true
	}

	isReddit := false
	if strings.HasPrefix( url, "https://www.reddit.com/r/") || strings.HasPrefix(url, "http://www.reddit.com/r/") {
		isReddit = true
	}

	rssfeed := RSSFeed{ID: GetUUID(), URL: url, Title: title, Description: feed.Description,
		Created: time.Now().String(), Updated: feed.Updated, Published: feed.Published,
		ChannelID: channel, Link: feed.Link, Author: author, Twitter: isTwitter, Reddit: isReddit}

	fmt.Println("Adding RSS Feed to DB: " + url + " Channel: " + channel)

	h.AddToDB(rssfeed)
	return nil
}


func (h *RSS) UpdateLastRun(lasttime time.Time, rssfeed RSSFeed) (err error){

	rssfeed.LastRun = lasttime

	db := h.db.rawdb.From("RSS")
	err = db.Update(&rssfeed)
	if err != nil {
		return err
	}
	return nil
}


func (h *RSS) UpdatePosts(rssfeed RSSFeed) (err error) {

	lastitem := rssfeed.LastItem

	for _, post := range rssfeed.Posts {
		if post == lastitem {
			return nil
		}
	}

	rssfeed.Posts = append(rssfeed.Posts, lastitem)

	db := h.db.rawdb.From("RSS")
	err = db.Update(&rssfeed)
	if err != nil {
		return err
	}

	return nil
}


func (h *RSS) GetLatestItem(url string, channel string) (rssitem RSSItem, err error) {

	rssfeed, err := h.GetFromDB(url, channel)
	if err != nil {
		return rssitem, err
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseUserAgentURL(rssfeed.URL, "Dual_Universe_Discord_Bot/1.0")
	if err != nil {
		return rssitem, err
	}


	if len(feed.Items) > 0 {

		if rssfeed.Twitter {

			feedAuthor := strings.TrimPrefix(rssfeed.Author, "https://twitter.com/")

			for _, item := range feed.Items {

				author := strings.Split(strings.TrimPrefix(item.Link,"https://twitter.com/"), "/")[0]

				if author == feedAuthor {
					rssitem.Twitter = true
					rssitem.Author = "@"+feedAuthor
					rssitem.Link = item.Link
					rssitem.Title = item.Title
					rssitem.Published = item.Published
					rssitem.Content = item.Content
					rssitem.Description = item.Description

					rssfeed.LastItem = item.Link
					h.UpdateLastRun(time.Now(), rssfeed)

					return rssitem, nil
				}

				h.UpdateLastRun(time.Now(), rssfeed)
				return rssitem, nil
			}
		}
		if rssfeed.Reddit {

			// The first two items on Reddit's RSS will be stickied posts, weird.
			item := feed.Items[2]
			rssitem.Twitter = false
			rssitem.Reddit = true
			rssitem.Author = item.Author.Name
			rssitem.Link = item.Link
			rssitem.Title = item.Title
			rssitem.Published = item.Published
			rssitem.Content = item.Content
			rssitem.Description = item.Description

			rssfeed.LastItem = item.Link
			h.UpdateLastRun(time.Now(), rssfeed)

			return rssitem, nil
		}

		item := feed.Items[0]
		rssitem.Twitter = false
		rssitem.Author = item.Author.Name
		rssitem.Link = item.Link
		rssitem.Title = item.Title
		rssitem.Published = item.Published
		rssitem.Content = item.Content
		rssitem.Description = item.Description

		rssfeed.LastItem = item.Link
		h.UpdateLastRun(time.Now(), rssfeed)

		return rssitem, nil

	}

	h.UpdateLastRun(time.Now(), rssfeed)
	return rssitem, errors.New("No items found!")
}
