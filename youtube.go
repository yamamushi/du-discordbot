package main

import (
	"fmt"
	"time"
	"os"
	"github.com/rylio/ytdl"
	"errors"
	"sync"
)

/*

This is a set of helper functions for parsing youtube links and id's into mp4 files for later conversion

 */

type YoutubeInterface struct {

	conf *Config
	querylocker sync.RWMutex
	playlistlocker sync.RWMutex
	db *DBHandler

}


type YoutubeRecord struct {

	ID	int `storm:"id,increment"`
	VideoID string `storm:"unique"`
	Title string
	Genre string  `storm:"index"`
	UserID string `storm:"index"`
	Author string `storm:"index"`

}


// DB Functions
// Records are stored in playlist groups

func (h *YoutubeInterface) AddToDB(record YoutubeRecord, playlist string) (err error){
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == ""{
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.Save(&record)
	return err
}


func (h *YoutubeInterface) RemoveFromDB(record YoutubeRecord, playlist string) (err error){
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == ""{
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.Remove(&record)
	return err
}


func (h *YoutubeInterface) GetFromDB(VideoID string, playlist string) (record YoutubeRecord, err error){
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == ""{
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.One("VideoID", VideoID, &record)
	if err != nil {
		return record, err
	}
	return record, nil
}


func (h *YoutubeInterface) GetDB(playlist string) (records []YoutubeRecord, err error){
	h.querylocker.Lock()
	defer h.querylocker.Unlock()
	if playlist == ""{
		playlist = "default"
	}

	db := h.db.rawdb.From("YouTube").From(playlist)
	err = db.All(&records)
	if err != nil{
		return records, err
	}
	return records, nil
}






// Download functions and helper functions

func (h *YoutubeInterface) AddToPlaylist(videoid string, userid string, playlist string, genre string)(err error){
	vid, err := h.CheckVideo(videoid)
	if err != nil {
		return err
	}

	record := YoutubeRecord{VideoID:vid.ID, Title: vid.Title, UserID: userid, Author: vid.Author, Genre: genre}
	err = h.AddToDB(record, playlist)
	if err != nil {
		return err
	}

	return nil
}



func (h *YoutubeInterface) RemoveFromPlaylist(videoid string, playlist string)(err error){

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




func (h *YoutubeInterface) DownloadYoutube(url string)(err error){

	vid, err := ytdl.GetVideoInfo(url)
	if err != nil{
		fmt.Println(err.Error())
		return err
	}

	if vid.Duration*time.Second > time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration))*time.Second {
		duration := vid.Duration.String()
		maxduration := time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration)).String()
		return errors.New("Video duration "+duration+" exceeds "+maxduration+" minutes!")
	}

	formats := vid.Formats
	file, _ := os.Create("tmp/"+vid.ID+".mp4")
	defer file.Close()
	if err = vid.Download(formats[1], file); err != nil {
		return err
	}

	return nil
}


func (h *YoutubeInterface) CheckVideo(url string)(info *ytdl.VideoInfo, err error){
	vid, err := ytdl.GetVideoInfo(url)
	if err != nil{
		return vid, err
	}

	if vid.Duration == 0 {
		return vid, errors.New("Invalid Video Link")
	}

	if vid.Duration*time.Second > time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration))*time.Second {

		duration := vid.Duration.String()
		maxduration := time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration)).String()
		return vid, errors.New("Video duration "+duration+" exceeds "+maxduration+" minutes!")

	}
	return vid, nil
}


func (h *YoutubeInterface) GetVideoObject(url string)(vid *ytdl.VideoInfo, err error) {

	vid, err = ytdl.GetVideoInfo(url)
	if err != nil{
		fmt.Println(err.Error())
		return nil, err
	}
	return vid, nil
}




func (h *YoutubeInterface) GetVideoID(url string)(id string, err error){
	vid, err := ytdl.GetVideoInfo(url)
	if err != nil{
		fmt.Println(err.Error())
		return "", err
	}

	if vid.Duration*time.Second > time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration))*time.Second {
		duration := vid.Duration.String()
		maxduration := time.Duration(time.Minute*time.Duration(h.conf.DUBotConfig.MaxAudioDuration)).String()
		return "", errors.New("Video duration "+duration+" exceeds "+maxduration+" minutes!")
	}
	return vid.ID, nil
}





