package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"

	//"fmt"
	"strings"
	"os"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"time"
	"bytes"
	"sync"
)

type StatsHandler struct {
	registry   	*CommandRegistry
	db         	*DBHandler
	statsdb 	*StatsDB
	conf 		*Config

	trackerlocker sync.RWMutex
	messageCount int
	activeusers []ActiveUser
}

type ActiveUser struct {
	UserID string
}

func (h *StatsHandler) Init(){
	h.statsdb = &StatsDB{db: h.db}
	h.RegisterCommands()
	CreateDirIfNotExist("./stats")

}

// RegisterCommands function
func (h *StatsHandler) RegisterCommands() (err error) {
	h.registry.Register("stats", "Manage discord statistics metrics", "")
	return nil
}

func (h *StatsHandler) Tracker(s *discordgo.Session, m *discordgo.MessageCreate){
	h.trackerlocker.Lock()
	defer h.trackerlocker.Unlock()
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	found := false
	for _, user := range h.activeusers {
		if user.UserID == m.Author.ID {
			found = true
		}
	}
	if !found {
		h.activeusers = append(h.activeusers, ActiveUser{UserID: m.Author.ID})
	}
	h.messageCount = h.messageCount + 1
}

func (h *StatsHandler) StatsWriter(s *discordgo.Session){
	for true {
		time.Sleep(time.Duration(time.Minute * 60))
		record, err := h.GetCurrentStats(s)
		if err == nil {
			h.statsdb.AddStatToDB(record)
			//fmt.Println("Added record to database")
			//fmt.Println(record.Date + " " + strconv.Itoa(record.TotalUsers) + " " + strconv.Itoa(record.OnlineUsers) + " " + strconv.Itoa(record.IdleUsers))
			//fmt.Println(strconv.Itoa(record.GamingUsers) + " " + strconv.Itoa(record.VoiceUsers) + " " + strconv.Itoa(record.MessageCount) + " " + strconv.Itoa(record.ActiveUserCount))
		}
	}
}

func (h *StatsHandler) GetCurrentStats(s *discordgo.Session) (record StatRecord, err error){
	h.trackerlocker.Lock()
	defer h.trackerlocker.Unlock()

	guild := s.State.Guilds[0]

	record.Date = time.Now().Format("2006-01-02 15:04:05")

	totalUserCount, err := h.GetUserCount(guild)
	if err != nil {
		return record, err
	}
	record.TotalUsers = totalUserCount

	onlineUserCount, err := h.GetOnlineCount(guild)
	if err != nil {
		return record, err
	}
	record.OnlineUsers = onlineUserCount

	idleUserCount, err := h.GetIdleCount(guild)
	if err != nil {
		return record, err
	}
	record.IdleUsers = idleUserCount

	gamingUserCount, err := h.GetGamingCount(guild)
	if err != nil {
		return record, err
	}
	record.GamingUsers = gamingUserCount

	voiceUserCount, err := h.GetVoiceCount(guild)
	if err != nil {
		return record, err
	}
	record.VoiceUsers = voiceUserCount

	record.MessageCount = h.messageCount
	h.messageCount = 0

	record.ActiveUserCount = len(h.activeusers)
	h.activeusers = []ActiveUser{}
	return record, nil
}


func (h *StatsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate){

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"stats") {
		if h.registry.CheckPermission("stats", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				//fmt.Println("error retrieving user:" + m.Author.ID)
				return
			}

			if user.Moderator {
				h.ParseCommand(command, s, m)
			}
		}
	}
}

func (h *StatsHandler) ParseCommand(commandarray []string, s *discordgo.Session, m *discordgo.MessageCreate){

	commandarray = RemoveStringFromSlice(commandarray, commandarray[0])
	if len(commandarray) == 0 {
		s.ChannelMessageSend(m.ChannelID, "The stats command is used to retrieve various statistics about " +
			"the discord server. You may manually load stats, unload stats, display graphs and various metrics about" +
			" the users on this discord with this command.")
		return
	}
	command, payload := SplitPayload(commandarray)

	if command == "help" {
		s.ChannelMessageSend(m.ChannelID, "The stats command is used to retrieve various statistics about " +
			"the discord server. You may manually load stats, unload stats, display graphs and various metrics about" +
				" the users on this discord with this command.")
		return
	}
	if command == "loadfromdisk" {
		db := h.db.rawdb.From("Users")
		var user User
		err := db.One("ID", m.Author.ID, &user)
		if err != nil {
			//fmt.Println("error retrieving user:" + m.Author.ID)
			return
		}
		if !user.Owner {
			s.ChannelMessageSend(m.ChannelID, "https://www.tenor.co/tewf.gif ")
			return
		}
		if len(payload) != 1 {
			s.ChannelMessageSend(m.ChannelID, "https://www.tenor.co/tewf.gif ")
			return
		}
		count, err := h.LoadFromDisk(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not load from disk: " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "DB updated from file: " + payload[0] + " added " + strconv.Itoa(count) + " records.")
		return
	}
	if command == "backer-count" {
		err := h.BackerPieChart(s,m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: " + err.Error())
			return
		}
		return
	}
	if command == "user-count" {
		err := h.UserCountChart(s,m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: " + err.Error())
			return
		}
		return
	}
}

func (h *StatsHandler) LoadFromDisk(path string) (count int, err error){
	if _, err := os.Stat(path); err == nil {
		//fmt.Println("Unmarshalling file...")
		raw, err := ioutil.ReadFile(path)
		if err != nil {
			return 0, err
		}
		var records []StatRecord
		err = json.Unmarshal(raw, &records)
		if err != nil {
			return 0, err
		}

		for _, record := range records {
			err = h.statsdb.AddStatToDB(record)
			if err != nil {
				return 0, err
			}
		}
		return len(records), nil
	}
	return 0, err
}

func (h *StatsHandler) BackerPieChart(s *discordgo.Session, m *discordgo.MessageCreate) (err error){

	members, err := GetMemberList(s)
	if err != nil {
		return err
	}

	// We are only concerned with gathering metrics on who has NDA access here
	// We can gather full forum authorized metrics in another function
	KyriumBacker := 0
	DiamondBacker := 0
	EmeraldBacker := 0
	RubyBacker := 0
	SapphireBacker := 0
	GoldBacker := 0
	PatronBacker := 0
	//ATVMember := 0

	for _, member := range members {
		for _, roleID := range member.Roles {
			if roleID == h.conf.RolesConfig.PatronRoleID {
				PatronBacker = PatronBacker+1
			}
			/*if roleID == h.conf.RolesConfig.ATVRoleID {
				ATVMember = ATVMember+1
			}*/
			if roleID == h.conf.RolesConfig.GoldRoleID {
				GoldBacker = GoldBacker+1
			}
			if roleID == h.conf.RolesConfig.SapphireRoleID {
				SapphireBacker = SapphireBacker+1
			}
			if roleID == h.conf.RolesConfig.RubyRoleID {
				RubyBacker = RubyBacker+1
			}
			if roleID == h.conf.RolesConfig.EmeraldRoleID {
				EmeraldBacker = EmeraldBacker+1
			}
			if roleID == h.conf.RolesConfig.DiamondRoleID {
				DiamondBacker = DiamondBacker+1
			}
			if roleID == h.conf.RolesConfig.KyriumRoleID {
				KyriumBacker = KyriumBacker+1
			}
		}
	}

	pieValues := []chart.Value{
		{Value: float64(PatronBacker), Label: "Patron - "+strconv.Itoa(PatronBacker)},
		//{Value: float64(ATVMember), Label: "ATV"+strconv.Itoa(PatronBacker)},
		{Value: float64(GoldBacker), Label: "Gold - "+strconv.Itoa(GoldBacker)},
		{Value: float64(SapphireBacker), Label: "Sapphire - "+strconv.Itoa(SapphireBacker)},
		{Value: float64(RubyBacker), Label: "Ruby - "+strconv.Itoa(RubyBacker)},
		{Value: float64(EmeraldBacker), Label: "Emerald - "+strconv.Itoa(EmeraldBacker)},
		{Value: float64(DiamondBacker), Label: "Diamond - "+strconv.Itoa(DiamondBacker)},
		{Value: float64(KyriumBacker), Label: "Kyrium - "+strconv.Itoa(KyriumBacker)},
	}

	style := chart.Style{ }
	style.FillColor = drawing.Color{131,135,142, 0 }
	pie := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: pieValues,
		Background: style,
	}

	filename := "./stats/PieChart-"+time.Now().Format("Jan 2 15-04-05")+".png"
	var b bytes.Buffer
	pie.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Backer Distribution ", s, m)
	if err != nil {
		return err
	}
	return nil
}

func (h *StatsHandler) WriteAndSend(path string, b bytes.Buffer, message string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	err = h.WriteImageFile(path, b)
	if err != nil {
		return err
	}
	// Now write the output to discord
	err = SendFileToChannel(path, message, s, m)
	if err != nil {
		return err
	}
	// Now we cleanup our temporary file
	os.Remove(path)
	return nil
}

func (h *StatsHandler) WriteImageFile(path string, b bytes.Buffer) (err error) {
	// open output file
	outputfile, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer outputfile.Close()

	// Write the output
	_, err = outputfile.Write(b.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (h *StatsHandler) UserCountChart(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	records, err := h.statsdb.GetFullDB()
	combinedRecords, err := h.CombineRecordsByDate(records)
	if err != nil {
		return err
	}
	//fmt.Println("Combined Record Count: " + strconv.Itoa(len(combinedRecords)))

	dateValues := []time.Time{}
	totalUsers := []float64{}

	for _, record := range combinedRecords {
		date, err := time.Parse("2006-01-02", record.Date)
		if err != nil {
			return err
		}
		dateValues = append(dateValues, date)
		totalUsers = append (totalUsers, float64(record.TotalUsers))
	}

	style := chart.Style{ }
	style.FillColor = drawing.Color{131,135,142, 0 }
	style.Show = true
	style.StrokeColor = chart.GetDefaultColor(0).WithAlpha(64)
	style.FillColor = chart.GetDefaultColor(0).WithAlpha(64)
	graph := chart.Chart{
		Title: "Total User Count",
		Background: chart.StyleShow(),
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateValues,
				YValues: totalUsers,
			},
		},
		XAxis: chart.XAxis{
			Name:      "Dates",
			NameStyle: chart.StyleShow(),
			Style:     style,
		},
		YAxis: chart.YAxis{
			Name:      "Users",
			NameStyle: chart.StyleShow(),
			Style:     style,
			ValueFormatter: h.FloatNormalize,
		},
	}

	filename := "./stats/TotalUsersChart-"+time.Now().Format("Jan 2 15-04-05")+".png"
	var b bytes.Buffer
	graph.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Total User Count ", s, m)
	if err != nil {
		return err
	}
	return nil
}


// TimeValueFormatter is a ValueFormatter for timestamps.
func (h *StatsHandler) FloatNormalize(v interface{}) string {
	if _, isTyped := v.(float64); isTyped {
		return strconv.Itoa(int(v.(float64)))
	}
	return ""
}

func (h *StatsHandler) CombineRecordsByDate(records []StatRecord) (combined []StatRecord, err error) {
	for _, record := range records {
		dateSplit := strings.Split(record.Date, " ")
		record.Date = dateSplit[0]

		if len(combined) < 1 {
			combined = append(combined, record)
		} else {
			found := false
			for _, tmpRecord := range combined {
				if record.Date == tmpRecord.Date {
					found = true
					tmpRecord.Engagement = tmpRecord.Engagement + record.Engagement
					tmpRecord.MessageCount = tmpRecord.MessageCount + record.MessageCount
					tmpRecord.VoiceUsers = tmpRecord.VoiceUsers + record.VoiceUsers
					tmpRecord.GamingUsers = tmpRecord.GamingUsers + record.GamingUsers
					tmpRecord.IdleUsers = tmpRecord.IdleUsers + record.IdleUsers
					tmpRecord.OnlineUsers = tmpRecord.OnlineUsers + record.OnlineUsers
					tmpRecord.TotalUsers = tmpRecord.TotalUsers + record.TotalUsers
				}
			}
			if !found {
				combined = append(combined, record)
			}
		}
	}

	return combined, nil
}

func (h *StatsHandler) GetUserCount(guild *discordgo.Guild) (count int, err error) {
	return guild.MemberCount, nil
}

func (h *StatsHandler) GetOnlineCount(guild *discordgo.Guild) (count int, err error) {
	for _, presence := range guild.Presences {
		if presence.Status == "online" {
			count = count + 1
		}
	}
	return count, nil
}

func (h *StatsHandler) GetIdleCount(guild *discordgo.Guild) (count int, err error) {
	for _, presence := range guild.Presences {
		if presence.Status == "idle" {
			count = count + 1
		}
	}
	return count, nil
}

func (h *StatsHandler) GetGamingCount(guild *discordgo.Guild) (count int, err error) {
	for _, presence := range guild.Presences {
		if presence.Game != nil {
			if presence.Game.State != "" {
				count = count + 1
			}
		}
	}
	return count, nil
}

func (h *StatsHandler) GetVoiceCount(guild *discordgo.Guild) (count int, err error) {
	return len(guild.VoiceStates), nil
}