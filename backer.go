package main

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"strings"
	"time"

	//"fmt"
	//"strconv"
	"sync"
)

type BackerInterface struct {
	db *DBHandler
	querylocker sync.Mutex
	conf *Config
}

type BackerRecord struct {
	UserID       string `storm:"id",json:"userid"`
	HashedID     string `json:"hashedid"`
	BackerStatus string `json:"backerstatus"`
	ForumProfile string `json:"forumprofile"`
	ATV          string `json:"atv"`
	PreAlpha     string `json:"prealpha"`
	Alpha        string `json:"alpha"`
	Validated    int    `json:"validated"`
}

// SaveRecordToDB function
func (h *BackerInterface) SaveRecordToDB(record BackerRecord) (err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		//log.Println("Could not connect to mongo: ", err.Error())
		return err
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("duauthbot").C(h.conf.DBConfig.BackerRecordColumn)

	_, err = c.UpsertId(record.UserID, record)
	return err
}

// NewPlayerRecord function
func (h *BackerInterface) NewPlayerRecord(userid string) (err error) {

	hashedid := h.HashUserID(userid)
	record := BackerRecord{UserID: userid, HashedID: hashedid, Validated: 0, BackerStatus: "nil"}
	err = h.SaveRecordToDB(record)
	return err

}

// GetRecordFromDB function
func (h *BackerInterface) GetRecordFromDB(userid string) (record BackerRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		//log.Println("Could not connect to mongo: ", err.Error())
		return record, err
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("duauthbot").C(h.conf.DBConfig.BackerRecordColumn)

	userrecord := BackerRecord{}
	err = c.Find(bson.M{"userid": userid}).One(&userrecord)
	return userrecord, err
}

// BackerInterface function
func (h *BackerInterface) GetAllBackers() (records []BackerRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		//log.Println("Could not connect to mongo: ", err.Error())
		return records, err
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("duauthbot").C(h.conf.DBConfig.BackerRecordColumn)

	err = c.Find(bson.M{}).All(&records)
	return records, err
}

// BackerInterface function
func (h *BackerInterface) GetAllBackersDeprecated() (records []BackerRecord, err error) {
	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	db := h.db.rawdb.From("DiscordAuth")
	err = db.All(&records)
	if err != nil {
		return records, err
	}

	return records, nil
}

// GetRecordFromDB function
func (h *BackerInterface) UniqueProfileCheck(userid string, profileurl string) (err error) {

	h.querylocker.Lock()
	defer h.querylocker.Unlock()

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{h.conf.DBConfig.MongoHost},
		Timeout:  30 * time.Second,
		Database: h.conf.DBConfig.MongoDB,
		Username: h.conf.DBConfig.MongoUser,
		Password: h.conf.DBConfig.MongoPass,
	}

	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		//log.Println("Could not connect to mongo: ", err.Error())
		return err
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("duauthbot").C(h.conf.DBConfig.BackerRecordColumn)

	userrecords := []BackerRecord{}
	err = c.Find(bson.M{}).All(&userrecords)
	if err != nil {
		if err.Error() == "not found" {
			return nil
		}
		return err
	}

	if len(userrecords) < 1 {
		return nil
	}

	for _, userrecord := range userrecords {
		if userrecord.UserID != userid {
			return errors.New("forum profile already assigned to userid: " + userrecord.UserID + " please contact an admin")
		}
	}

	return nil
}

// UserHasRecord function
func (h *BackerInterface) UserHasRecord(userid string) bool {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return false
	}

	if record.UserID == "" {
		return false
	}
	return true
}

func (h *BackerInterface) UserValidated(userid string) bool {

	if !h.UserHasRecord(userid) {
		err := h.NewPlayerRecord(userid)
		if err != nil {
			return false
		}
	}

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return false
	}
	if record.Validated == 0 {
		return false
	}
	return true
}

// SetForumProfile function
func (h *BackerInterface) SetValidatedStatus(userid string, validated int) (err error) {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.Validated = validated
	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

// UserHasRecord function
func (h *BackerInterface) GetBackerStatus(userid string) (status string, err error) {
	if !h.UserHasRecord(userid) {
		return "", errors.New("Error: No User Record Exists!")
	}
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return "", err
	}
	return record.BackerStatus, nil
}

// SetForumProfile function
func (h *BackerInterface) SetBackerStatus(userid string, backerstatus string) (err error) {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.BackerStatus = backerstatus
	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

// GetATVStatus function
func (h *BackerInterface) GetATVStatus(userid string) (status string, err error) {
	if !h.UserHasRecord(userid) {
		return "", errors.New("Error: No User Record Exists!")
	}
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return "", err
	}
	return record.ATV, nil
}

// GetATVStatus function
func (h *BackerInterface) SetATVStatus(userid string, atvstatus string) (err error) {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.ATV = atvstatus
	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

// GetATVStatus function
func (h *BackerInterface) GetPreAlphaStatus(userid string) (status string, err error) {
	if !h.UserHasRecord(userid) {
		return "", errors.New("Error: No User Record Exists!")
	}
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return "", err
	}
	return record.PreAlpha, nil
}

// GetATVStatus function
func (h *BackerInterface) SetPreAlphaStatus(userid string, prealphastatus string) (err error) {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.PreAlpha = prealphastatus
	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

// UserHasRecord function
func (h *BackerInterface) GetForumProfile(userid string) (profileurl string, err error) {
	if !h.UserHasRecord(userid) {
		return "", errors.New("Error: No User Record Exists!")
	}
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return "", err
	}
	return record.ForumProfile, nil
}

// SetForumProfile function
func (h *BackerInterface) SetForumProfile(userid string, profileurl string) (err error) {

	//fmt.Println("Unique Profile Check")
	err = h.UniqueProfileCheck(userid, profileurl)
	if err != nil {
		return err
	}

	//fmt.Println("Getting DB Record")
	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	record.ForumProfile = profileurl
	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}
	return nil
}

func (h *BackerInterface) HashUserID(userid string) string {

	hasher := sha256.New()
	io.WriteString(hasher, userid)
	sha256hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha256hash

}

func (h *BackerInterface) ResetUser(userid string) error {

	/* if !h.UserHasRecord(userid){
		return errors.New("Error: No User Record Exists!")
	} */

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}
	record.Validated = 0
	record.ATV = "false"
	record.ForumProfile = ""
	record.PreAlpha = "false"

	err = h.SaveRecordToDB(record)
	if err != nil {
		return err
	}

	//fmt.Println(record.Validated)
	return nil
}

func (h *BackerInterface) ForumAuth(url string, userid string) (err error) {

	if !h.UserHasRecord(userid) {
		err := h.NewPlayerRecord(userid)
		if err != nil {
			fmt.Println("Error: " + err.Error())
			return err
		}
	}

	if !strings.HasPrefix(url, "https://board.dualthegame.com/index.php?/profile/") {
		return errors.New("expected url from https://board.dualthegame.com/index.php?/profile/")
	}

	//fmt.Println("Setting Forum Profile")
	err = h.SetForumProfile(userid, url)
	if err != nil {
		return err
	}

	//fmt.Println("Checking User Validation")
	err = h.CheckUserValidation(userid)
	if err != nil {
		return err
	}

	//fmt.Println("Checking User Status")
	err = h.CheckStatus(userid)
	if err != nil {
		return err
	}

	//fmt.Println("Checking Backer Status")
	backerstatus, err := h.GetBackerStatus(userid)
	if err != nil {
		return err
	}

	atvstatus, err := h.GetATVStatus(userid)
	if err != nil {
		return err
	}

	prealphastatus, err := h.GetPreAlphaStatus(userid)
	if err != nil {
		return err
	}

	if backerstatus != "" || atvstatus == "true" || prealphastatus == "true" {
		err = h.SetValidatedStatus(userid, 1)
		if err != nil {
			//fmt.Println("Setting validated status")
			return err
		}
	}
	return nil
}

func (h *BackerInterface) CheckUserValidation(userid string) (err error) {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	validation, err := h.GetValidationString(record)
	if err != nil {
		return err
	}

	if validation != record.HashedID {
		return errors.New("invalid user validation token found: " + validation + " expected: " + record.HashedID)
	}

	return nil
}

// URL Interaction Functions

func (h *BackerInterface) GetValidationString(record BackerRecord) (validation string, err error) {

	resp, err := soup.Get(record.ForumProfile) // Append page=1000 so we get the last page
	if err != nil {
		//fmt.Println("Could not retreive page: " + record.ForumProfile)
		return "", err
	}

	doc := soup.HTMLParse(resp)
	activityStream := doc.Find("div", "class", "ipsTabs_panels ipsPad_double ipsAreaBackground_reset").FindAll("li")

	if len(activityStream) > 0 {

		for _, activityitem := range activityStream {
			commenters := activityitem.Find("div", "class", "ipsStreamItem_container").FindAll("div")
			if len(commenters) > 0 {
				for _, comments := range commenters {
					content := comments.FindAll("span")
					if len(content) > 0 {
						if content[0].Attrs()["title"] == "Status Update" {
							commenterurls := comments.FindAll("a")
							if len(commenterurls) > 0 {
								if commenterurls[0].Attrs()["href"] == record.ForumProfile {
									commentp := activityitem.Find("div", "class", "ipsStreamItem_snippet").FindAll("p")
									if len(commentp) > 0 {
										message := commentp[0].Text()
										//fmt.Print(message)
										//fmt.Println(strconv.Itoa(len(commentp)))
										if strings.Contains(message, "discordauth") {
											//fmt.Println(message)
											fields := strings.Split(message, ":")
											if len(fields) == 2 {
												checksum := strings.TrimSuffix(fields[1], "\n")
												return checksum, nil
											}
										}
									}

									span := activityitem.Find("div", "class", "ipsStreamItem_snippet").FindAll("span")
									if len(span) > 0 {
										message := span[0].Text()
										//fmt.Print(message)
										//fmt.Println(strconv.Itoa(len(span)))
										if strings.Contains(message, "discordauth") {
											fields := strings.Split(message, ":")
											if len(fields) == 2 {
												checksum := strings.TrimSuffix(fields[1], "\n")
												return checksum, nil
											}
										}
									}

									code := activityitem.Find("div", "class", "ipsStreamItem_snippet").FindAll("code")
									if len(code) > 0 {
										message := code[0].Text()
										//fmt.Print(message)
										//fmt.Println(strconv.Itoa(len(code)))
										if strings.Contains(message, "discordauth") {
											fields := strings.Split(message, ":")
											if len(fields) == 2 {
												checksum := strings.TrimSuffix(fields[1], "\n")
												return checksum, nil
											}
										}
									}

								}
							}
						}
					}
				}
			}
		}
	}

	return "", errors.New("could not retrieve validation string")
}

func (h *BackerInterface) CheckStatus(userid string) (err error) {

	record, err := h.GetRecordFromDB(userid)
	if err != nil {
		return err
	}

	backerstatus := h.GetBackerString(record)
	if backerstatus != "" {
		err = h.SetBackerStatus(userid, backerstatus)
		if err != nil {
			return err
		}
	}

	if h.GetATVString(record) {
		err = h.SetATVStatus(userid, "true")
		if err != nil {
			return err
		}
		err = h.SetPreAlphaStatus(userid, "true")
		if err != nil {
			return err
		}
	}

	if h.GetPreAlphaString(record) {
		err = h.SetPreAlphaStatus(userid, "true")
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *BackerInterface) GetBackerString(record BackerRecord) (status string) {

	resp, err := soup.Get(record.ForumProfile) // Append page=1000 so we get the last page
	if err != nil {
		//fmt.Println("Could not retrieve page: " + record.ForumProfile)
		return ""
	}

	doc := soup.HTMLParse(resp)
	profile_info := doc.FindAll("div", "class", "ipsWidget ipsWidget_vertical cProfileSidebarBlock ipsBox ipsSpacer_bottom")

	if len(profile_info) > 0 {
		for _, field := range profile_info {
			//fmt.Println(field.Attrs())
			inner_items := field.Find("div", "class", "ipsWidget_inner ipsPad").FindAll("li")

			if len(inner_items) > 0 {
				for _, pad := range inner_items {
					pad_titles := pad.Find("span", "class", "ipsDataItem_generic ipsDataItem_size3 ipsType_break").FindAll("strong")
					pad_contents := pad.FindAll("div", "class", "ipsType_break ipsContained")
					if len(pad_titles) > 0 {
						for _, title := range pad_titles {
							if title.Text() == "backer_title" {
								if len(pad_contents) > 0 {
									//fmt.Println(pad_contents[0].Text())
									return pad_contents[0].Text()
								}
							}
						}
					}
				}
			}
		}
	}

	return ""
}

func (h *BackerInterface) GetPreAlphaString(record BackerRecord) (status bool) {

	resp, err := soup.Get(record.ForumProfile) // Append page=1000 so we get the last page
	if err != nil {
		//fmt.Println("Could not retreive page: " + record.ForumProfile)
		return false
	}

	doc := soup.HTMLParse(resp)
	profile_header := doc.FindAll("header", "data-role", "profileHeader")

	if len(profile_header) > 0 {
		for _, headers := range profile_header {
			bar_text := headers.FindAll("span", "class", "ipsPageHead_barText")
			if len(bar_text) > 0 {
				status := bar_text[0].FindAll("span")
				if len(status) > 0 {
					if status[0].Text() == "Pre-Alpha Tester" {
						return true
					}
				}
			}
		}
	}

	return false
}

func (h *BackerInterface) GetATVString(record BackerRecord) (status bool) {

	resp, err := soup.Get(record.ForumProfile) // Append page=1000 so we get the last page
	if err != nil {
		//fmt.Println("Could not retreive page: " + record.ForumProfile)
		return false
	}

	doc := soup.HTMLParse(resp)
	profile_header := doc.FindAll("header", "data-role", "profileHeader")

	if len(profile_header) > 0 {
		for _, headers := range profile_header {
			bar_text := headers.FindAll("span", "class", "ipsPageHead_barText")
			if len(bar_text) > 0 {
				status := bar_text[0].FindAll("span")
				if len(status) > 0 {
					if status[0].Text() == "Alpha Team Vanguard" {
						return true
					}
				}
			}
		}
	}

	return false
}
