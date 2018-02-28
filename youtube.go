package main

/*import (
	"errors"
	"fmt"
	"github.com/rylio/ytdl"
	"os"
	"sync"
	"time"
)

/*

This is a set of helper functions for parsing youtube links and id's into mp4 files for later conversion

/*
// YoutubeInterface is our way of interacting with YouTube, and the means by which we
// store retrieved video information in our database.
type YoutubeInterface struct {
	conf           *Config
	querylocker    sync.RWMutex
	playlistlocker sync.RWMutex
	db             *DBHandler
}

// YoutubeRecord is a struct that we pass into our database for storing video information.
type YoutubeRecord struct {
	ID      int    `storm:"id,increment"`
	VideoID string `storm:"unique"`
	Title   string
	Genre   string `storm:"index"`
	UserID  string `storm:"index"`
	Author  string `storm:"index"`
}

// DB Functions

// AddToDB will add the given YoutubeRecord to the provided playlist (which defaults to "default" if none is provided)
func (h *YoutubeInterface) AddToDB(record YoutubeRecord, playlist string) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == "" {
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.Save(&record)
	return err
}

// RemoveFromDB will remove the given YoutubeRecord from the provided playlist (which defaults to "default")
func (h *YoutubeInterface) RemoveFromDB(record YoutubeRecord, playlist string) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == "" {
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.DeleteStruct(&record)
	return err
}

// GetFromDB will retrieve the provided VideoID (which must be an ID not a URL) from the provided playlist (which defaults to "default")
func (h *YoutubeInterface) GetFromDB(VideoID string, playlist string) (record YoutubeRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == "" {
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.One("VideoID", VideoID, &record)
	if err != nil {
		return record, err
	}
	return record, nil
}

// GetDB will return the entire list of videos for the provided playlist (which defaults to "default")
func (h *YoutubeInterface) GetDB(playlist string) (records []YoutubeRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == "" {
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.All(&records)
	if err != nil {
		return records, err
	}
	return records, nil
}

// Download functions and helper functions

// AddToPlaylist will add a video's metadata to the provided playlist (which does not have a default entry), and insert it into
// The database for later retrieval. This does not store a copy of the video in the database.
// videoid can be a URL here, as CheckVideo will return the valid information (if the url is valid)
func (h *YoutubeInterface) AddToPlaylist(videoid string, userid string, playlist string, genre string) (err error) {
	vid, err := h.CheckVideo(videoid)
	if err != nil {
		return err
	}

	record := YoutubeRecord{VideoID: vid.ID, Title: vid.Title, UserID: userid, Author: vid.Author, Genre: genre}
	err = h.AddToDB(record, playlist)
	if err != nil {
		return err
	}

	return nil
}

// RemoveFromPlaylist will remove the provided video (based on the VideoID which must be a valid ID) from the
// provided playlist (which has no default).
func (h *YoutubeInterface) RemoveFromPlaylist(videoid string, playlist string) (err error) {

	record, err := h.GetFromDB(videoid, playlist)
	if err != nil {
		return err
	}

	err = h.RemoveFromDB(record, playlist)
	if err != nil {
		return err
	}

	return nil
}

// DownloadYoutube will download a video from the provided URL, as long as the duration does not exceed
// The max duration set in the configuration file before startup.
func (h *YoutubeInterface) DownloadYoutube(url string) (err error) {

	vid, err := ytdl.GetVideoInfo(url)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if vid.Duration*time.Second > time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration))*time.Second {
		duration := vid.Duration.String()
		maxduration := time.Duration(time.Minute * time.Duration(h.conf.DUBotConfig.MaxAudioDuration)).String()
		return errors.New("Video duration " + duration + " exceeds " + maxduration + " minutes!")
	}

	formats := vid.Formats
	file, _ := os.Create("tmp/" + vid.ID + ".mp4")
	defer file.Close()
	if err = vid.Download(formats[1], file); err != nil {
		return err
	}

	return nil
}

// CheckVideo will parse a URL to determine if it is valid for downloading, and to verify that the URL is in fact a valid one.
// If the duration is 0, we know we did not successfully retrieve a video.
func (h *YoutubeInterface) CheckVideo(url string) (info *ytdl.VideoInfo, err error) {
	vid, err := ytdl.GetVideoInfo(url)
	if err != nil {
		return vid, err
	}

	if vid.Duration == 0 {
		return vid, errors.New("Invalid Video Link")
	}

	if vid.Duration*time.Second > time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration))*time.Second {

		duration := vid.Duration.String()
		maxduration := time.Duration(time.Minute * time.Duration(h.conf.DUBotConfig.MaxAudioDuration)).String()
		return vid, errors.New("Video duration " + duration + " exceeds " + maxduration + " minutes!")

	}
	return vid, nil
}

// GetVideoObject will return the ytdl.VideoInfo struct for a URL, useful for grabbing information about a URL quickly.
func (h *YoutubeInterface) GetVideoObject(url string) (vid *ytdl.VideoInfo, err error) {

	vid, err = ytdl.GetVideoInfo(url)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return vid, nil
}

// GetVideoID will return the ID for a provided URL
func (h *YoutubeInterface) GetVideoID(url string) (id string, err error) {
	vid, err := ytdl.GetVideoInfo(url)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	if vid.Duration*time.Second > time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration))*time.Second {
		duration := vid.Duration.String()
		maxduration := time.Duration(time.Minute * time.Duration(h.conf.DUBotConfig.MaxAudioDuration)).String()
		return "", errors.New("Video duration " + duration + " exceeds " + maxduration + " minutes!")
	}
	return vid.ID, nil
}
*/
