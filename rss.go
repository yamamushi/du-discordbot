package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yamamushi/gofeed"
	"sync"
)

// RSS struct
type RSS struct {
	db          *DBHandler
	querylocker sync.RWMutex
}

// RSSFeed struct
type RSSFeed struct {
	ID          string `storm:"id"`
	Title       string
	URL         string    `storm:"index"`
	LastRun     time.Time // Not particularly needed, but may become useful
	Created     string
	Description string
	Author      string
	Updated     string
	Published   string
	ChannelID   string `storm:"index"` // Limit our feeds per channel
	Link        string
	Twitter     bool `storm:"index"`
	Reddit      bool
	Youtube     bool
	Forum       bool
	LastItem    string // URL of last item
	RepeatPosts bool
	Posts       []string // List of Posted URL's
}

// RSSItem struct
type RSSItem struct {
	Title       string
	Author      string
	Content     string
	Description string
	Published   string
	Link        string
	Twitter     bool
	Reddit      bool
	Youtube     bool
	Forum       bool
	Update      bool
}

// Validate function
func (h *RSS) Validate(url string) error {

	fp := gofeed.NewParser()
	_, err := fp.ParseURL(url)
	if err != nil {
		return err
	}

	return nil
}

// AddToDB function
func (h *RSS) AddToDB(rssfeed RSSFeed) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RSS")
	err = db.Save(&rssfeed)
	return err
}

// RemoveFromDB function
func (h *RSS) RemoveFromDB(rssfeed RSSFeed) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RSS")
	err = db.DeleteStruct(&rssfeed)
	return err
}

// GetFromDB function
func (h *RSS) GetFromDB(url string, channel string) (rssfeed RSSFeed, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

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

// GetDB function
func (h *RSS) GetDB() (rssfeeds []RSSFeed, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("RSS")
	err = db.All(&rssfeeds)
	if err != nil {
		return rssfeeds, err
	}
	return rssfeeds, nil
}

// GetChannel function
func (h *RSS) GetChannel(channel string) (rssfeeds []RSSFeed, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

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

// CheckDB function
func (h *RSS) CheckDB(url string, channel string) bool {

	_, err := h.GetFromDB(url, channel)
	if err != nil {
		return true
	}

	return false
}

// GetTitle function
func (h *RSS) GetTitle(url string, channel string) (title string, err error) {

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

// GetDescription function
func (h *RSS) GetDescription(url string, channel string) (description string, err error) {

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

// GetUpdated function
func (h *RSS) GetUpdated(url string, channel string) (updated string, err error) {

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

// GetPublished function
func (h *RSS) GetPublished(url string, channel string) (published string, err error) {

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

	rssfeed, err := h.GetRecordFromDB(url, channel)
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

// Subscribe function
func (h *RSS) Subscribe(url string, title string, channel string, repeatposts bool) (err error) {

	if !h.CheckDB(url, channel) {
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
	if strings.HasPrefix(url, "https://twitrss.me/") {
		author = strings.TrimPrefix(feed.Link, "https://twitter.com/")
		isTwitter = true
	}

	isReddit := false
	if strings.HasPrefix(url, "https://www.reddit.com/r/") || strings.HasPrefix(url, "http://www.reddit.com/r/") {
		isReddit = true
	}

	isYoutube := false
	if strings.HasPrefix(url, "https://www.youtube.com") || strings.HasPrefix(url, "http://www.youtube.com") {
		isYoutube = true
	}

	isForum := false
	if strings.HasPrefix(url, "https://board.dualthegame.com") || strings.HasPrefix(url, "http://board.dualthegame.com") {
		isForum = true
	}

	uuid, err := GetUUID()
	if err != nil {
		return errors.New("Fatal Error generating UUID: " + err.Error())
	}
	rssfeed := RSSFeed{ID: uuid, URL: url, Title: title, Description: feed.Description,
		Created: time.Now().String(), Updated: feed.Updated, Published: feed.Published,
		ChannelID: channel, Link: feed.Link, Author: author, Twitter: isTwitter, Reddit: isReddit,
		Youtube: isYoutube, Forum: isForum, RepeatPosts: repeatposts}

	//fmt.Println("Adding RSS Feed to DB: " + url + " Channel: " + channel)

	h.AddToDB(rssfeed)
	return nil
}

// Unsubscribe function
func (h *RSS) Unsubscribe(url string, channel string) (err error) {

	rssfeed, err := h.GetFromDB(url, channel)
	if err != nil {
		return err
	}

	err = h.RemoveFromDB(rssfeed)
	if err != nil {
		return err
	}

	fmt.Println("Removing RSS Feed from DB: " + url + " Channel: " + channel)

	return nil
}

// UpdateLastRun function
func (h *RSS) UpdateLastRun(lasttime time.Time, rssfeed RSSFeed) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	rssfeed.LastRun = lasttime

	db := h.db.rawdb.From("RSS")
	err = db.Update(&rssfeed)
	if err != nil {
		return err
	}
	return nil
}

// UpdatePosts function
func (h *RSS) UpdatePosts(rssfeed RSSFeed) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

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

// GetLatestItem function
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

			var i = 0
			if len(feed.Items) > 1 {
				pubone := strings.Split(feed.Items[0].Published, "+")
				firstItemPubDate, err := time.Parse("Mon _2 Jan 2006 15:04:05", strings.Replace(strings.TrimSpace(pubone[0]), ",", "", -1))
				if err != nil {
					return rssitem, err
				}
				pubtwo := strings.Split(feed.Items[1].Published, "+")
				secondItemPubDate, err := time.Parse("Mon _2 Jan 2006 15:04:05", strings.Replace(strings.TrimSpace(pubtwo[0]), ",", "", -1))
				if err != nil {
					return rssitem, err
				}
				if firstItemPubDate.Before(secondItemPubDate) {
					i = 1
				}
			}

			item := feed.Items[i]

			author := strings.Split(strings.TrimPrefix(item.Link, "https://twitter.com/"), "/")[0]

			if author == feedAuthor {
				rssitem.Twitter = true
				rssitem.Reddit = false
				rssitem.Youtube = false
				rssitem.Forum = false
				rssitem.Author = "@" + feedAuthor
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

		} else if rssfeed.Reddit {

			// The first two items on Reddit's RSS will be stickied posts, weird.
			item := feed.Items[1]

			rssitem.Twitter = false
			rssitem.Reddit = true
			rssitem.Youtube = false
			rssitem.Forum = false
			rssitem.Author = item.Author.Name
			rssitem.Link = item.Link
			rssitem.Title = item.Title
			rssitem.Published = item.Published
			rssitem.Content = item.Content
			rssitem.Description = item.Description

			rssfeed.LastItem = item.Link
			h.UpdateLastRun(time.Now(), rssfeed)

			return rssitem, nil
		} else if rssfeed.Youtube {

			item := feed.Items[0]
			rssitem.Twitter = false
			rssitem.Reddit = false
			rssitem.Youtube = true
			rssitem.Forum = false
			rssitem.Author = item.Author.Name
			rssitem.Link = item.Link
			rssitem.Title = item.Title
			rssitem.Published = item.Published
			rssitem.Content = item.Content
			rssitem.Description = item.Description

			rssfeed.LastItem = item.Link
			h.UpdateLastRun(time.Now(), rssfeed)

			return rssitem, nil
		} else if rssfeed.Forum {

			item := feed.Items[0]
			rssitem.Twitter = false
			rssitem.Reddit = false
			rssitem.Youtube = false
			rssitem.Forum = true
			rssitem.Author = ""
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
		rssitem.Reddit = false
		rssitem.Youtube = false
		rssitem.Forum = false
		if item.Author != nil {
			rssitem.Author = item.Author.Name
		} else {
			rssitem.Author = ""
		}
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
	return rssitem, errors.New("No items found for: " + url)
}
