package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"

	//"fmt"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StatsHandler struct {
	registry *CommandRegistry
	db       *DBHandler
	statsdb  *StatsDB
	conf     *Config

	trackerlocker sync.RWMutex
	messageCount  int
	activeusers   []ActiveUser
}

type ActiveUser struct {
	UserID string
}

func (h *StatsHandler) Init() {
	h.statsdb = &StatsDB{db: h.db}
	h.RegisterCommands()
	CreateDirIfNotExist("./stats")

}

// RegisterCommands function
func (h *StatsHandler) RegisterCommands() (err error) {
	h.registry.Register("stats", "Manage discord statistics metrics", "")
	return nil
}

func (h *StatsHandler) Tracker(s *discordgo.Session, m *discordgo.MessageCreate) {
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

func (h *StatsHandler) StatsWriter(s *discordgo.Session) {
	for true {
		time.Sleep(time.Duration(time.Second * 30))
		record, err := h.GetCurrentStats(s)
		if err == nil {
			err = h.statsdb.AddStatToDB(record)
			if err != nil {
				fmt.Println(err.Error())
			}
			h.messageCount = 0
			h.activeusers = []ActiveUser{}
			//fmt.Println("Added record to database")
			//fmt.Println(record.Date + " " + strconv.Itoa(record.TotalUsers) + " " + strconv.Itoa(record.OnlineUsers) + " " + strconv.Itoa(record.IdleUsers))
			//fmt.Println(strconv.Itoa(record.GamingUsers) + " " + strconv.Itoa(record.VoiceUsers) + " " + strconv.Itoa(record.MessageCount) + " " + strconv.Itoa(record.ActiveUserCount))
		}
		time.Sleep(time.Duration(time.Minute * h.conf.DUBotConfig.StatsTimeout))
	}
}

func (h *StatsHandler) GetCurrentStats(s *discordgo.Session) (record StatRecord, err error) {
	h.trackerlocker.Lock()
	defer h.trackerlocker.Unlock()

	guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
	if err != nil {
		return record, err
	}

	record.Date = time.Now().Format("2006-01-02 15:04:05")
	/*
		totalUserCount, err := h.GetUserCount(s)
		if err != nil {
			return record, err
		}
	*/

	record.TotalUsers = guild.MemberCount

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

	invisibleUserCount, err := h.GetInvisibleCount(guild)
	if err != nil {
		return record, err
	}
	record.InvisibleUsers = invisibleUserCount

	dndUserCount, err := h.GetDNDCount(guild)
	if err != nil {
		return record, err
	}
	record.DNDUsers = dndUserCount

	record.MessageCount = h.messageCount
	record.Engagement = len(h.activeusers)
	return record, nil
}

func (h *StatsHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

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

func (h *StatsHandler) ParseCommand(commandarray []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	commandarray = RemoveStringFromSlice(commandarray, commandarray[0])
	if len(commandarray) == 0 {
		s.ChannelMessageSend(m.ChannelID, "The stats command is used to retrieve various statistics about "+
			"the discord server. You may manually load stats, unload stats, display graphs and various metrics about"+
			" the users on this discord with this command.")
		return
	}
	command, payload := SplitPayload(commandarray)

	if command == "help" {
		s.ChannelMessageSend(m.ChannelID, "The stats command is used to retrieve various statistics about "+
			"the discord server. You may manually load stats, unload stats, display graphs and various metrics about"+
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
			s.ChannelMessageSend(m.ChannelID, "Could not load from disk: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "DB updated from file: "+payload[0]+" added "+strconv.Itoa(count)+" records.")
		return
	}
	if command == "backer-count" {
		err := h.BackerPieChart(s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
			return
		}
		return
	}
	if command == "nda-count" {
		err := h.NDAPieChart(s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
			return
		}
		return
	}
	if command == "user-count" {
		if len(payload) > 0 {
			days, err := strconv.Atoi(payload[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Flag should be an integer value")
				return
			}
			err = h.UserCountChart(days, s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
				return
			}
			return
		}
		err := h.UserCountChart(365, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
			return
		}
		return
	}
	if command == "daily-active" {
		if len(payload) > 0 {
			days, err := strconv.Atoi(payload[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Flag should be an integer value")
				return
			}
			err = h.DailyActiveChart(days, s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
				return
			}
			return
		}
		err := h.DailyActiveChart(30, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
			return
		}
		return
	}
	if command == "daily-online" {
		if len(payload) > 0 {
			days, err := strconv.Atoi(payload[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Flag should be an integer value")
				return
			}
			err = h.DailyOnlineChart(days, s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
				return
			}
			return
		}
		err := h.DailyOnlineChart(30, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
			return
		}
		return
	}
	if command == "daily-voice" {
		if len(payload) > 0 {
			days, err := strconv.Atoi(payload[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Flag should be an integer value")
				return
			}
			err = h.DailyVoiceChart(days, s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
				return
			}
			return
		}
		err := h.DailyVoiceChart(30, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
			return
		}
		return
	}
	if command == "daily-idle" {
		if len(payload) > 0 {
			days, err := strconv.Atoi(payload[0])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: Flag should be an integer value")
				return
			}
			err = h.DailyIdleChart(days, s, m)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
				return
			}
			return
		}
		err := h.DailyIdleChart(30, s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating chart: "+err.Error())
			return
		}
		return
	}
	if command == "show-invisible" {
		err := h.ShowInvisibleUsers(s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating list: "+err.Error())
			return
		}
		return
	}
	if command == "show-dnd" {
		err := h.ShowDNDUsers(s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error generating list: "+err.Error())
			return
		}
		return
	}
	if command == "repair-totals" {
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

		guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error retrieving guild: "+err.Error())
			return
		}
		err = h.RepairTotals(guild)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error repairing database: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Repair complete")
		return
	}
}

func (h *StatsHandler) LoadFromDisk(path string) (count int, err error) {
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

func (h *StatsHandler) BackerPieChart(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	members, err := GetMemberList(s, h.conf)
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
	SilverBacker := 0
	BronzeBacker := 0
	IronBacker := 0
	PatronBacker := 0
	SponsorBacker := 0
	ContributorBacker := 0
	BackerCount := 0
	//ATVMember := 0

	for _, member := range members {
		for _, roleID := range member.Roles {
			backer := false
			if roleID == h.conf.RolesConfig.ContributorRoleID {
				ContributorBacker = ContributorBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.SponsorRoleID {
				SponsorBacker = SponsorBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.PatronRoleID {
				PatronBacker = PatronBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.IronRoleID {
				IronBacker = IronBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.BronzeRoleID {
				BronzeBacker = BronzeBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.SilverRoleID {
				SilverBacker = SilverBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.GoldRoleID {
				GoldBacker = GoldBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.SapphireRoleID {
				SapphireBacker = SapphireBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.RubyRoleID {
				RubyBacker = RubyBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.EmeraldRoleID {
				EmeraldBacker = EmeraldBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.DiamondRoleID {
				DiamondBacker = DiamondBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.KyriumRoleID {
				KyriumBacker = KyriumBacker + 1
				backer = true
			}
			if backer {
				BackerCount = BackerCount + 1
			}
		}
	}
	PatronStyle := chart.Style{FillColor: drawing.Color{R: 214, G: 187, B: 32, A: 255}}
	SponsorStyle := chart.Style{FillColor: drawing.Color{R: 143, G: 143, B: 143, A: 255}}
	ContributorStyle := chart.Style{FillColor: drawing.Color{R: 252, G: 62, B: 99, A: 255}}
	IronStyle := chart.Style{FillColor: drawing.Color{R: 143, G: 143, B: 143, A: 255}}
	BronzeStyle := chart.Style{FillColor: drawing.Color{R: 148, G: 101, B: 6, A: 255}}
	SilverStyle := chart.Style{FillColor: drawing.Color{R: 222, G: 222, B: 222, A: 255}}
	GoldStyle := chart.Style{FillColor: drawing.Color{R: 217, G: 176, B: 80, A: 255}}
	SapphireStyle := chart.Style{FillColor: drawing.Color{R: 0, G: 102, B: 255, A: 255}}
	RubyStyle := chart.Style{FillColor: drawing.Color{R: 255, G: 3, B: 3, A: 255}}
	EmeraldStyle := chart.Style{FillColor: drawing.Color{R: 7, G: 230, B: 137, A: 255}}
	DiamondStyle := chart.Style{FillColor: drawing.Color{R: 255, G: 255, B: 255, A: 255}}
	KyriumStyle := chart.Style{FillColor: drawing.Color{R: 219, G: 219, B: 219, A: 255}}

	pieValues := []chart.Value{
		{Value: float64(ContributorBacker), Label: "Contributor - " + strconv.Itoa(ContributorBacker), Style: ContributorStyle},
		{Value: float64(SponsorBacker), Label: "Sponsor - " + strconv.Itoa(SponsorBacker), Style: SponsorStyle},
		{Value: float64(PatronBacker), Label: "Patron - " + strconv.Itoa(PatronBacker), Style: PatronStyle},
		{Value: float64(IronBacker), Label: "Iron - " + strconv.Itoa(IronBacker), Style: IronStyle},
		{Value: float64(BronzeBacker), Label: "Bronze - " + strconv.Itoa(BronzeBacker), Style: BronzeStyle},
		{Value: float64(SilverBacker), Label: "Silver - " + strconv.Itoa(SilverBacker), Style: SilverStyle},
		{Value: float64(GoldBacker), Label: "Gold - " + strconv.Itoa(GoldBacker), Style: GoldStyle},
		{Value: float64(SapphireBacker), Label: "Sapphire - " + strconv.Itoa(SapphireBacker), Style: SapphireStyle},
		{Value: float64(RubyBacker), Label: "Ruby - " + strconv.Itoa(RubyBacker), Style: RubyStyle},
		{Value: float64(EmeraldBacker), Label: "Emerald - " + strconv.Itoa(EmeraldBacker), Style: EmeraldStyle},
		{Value: float64(DiamondBacker), Label: "Diamond - " + strconv.Itoa(DiamondBacker), Style: DiamondStyle},
		{Value: float64(KyriumBacker), Label: "Kyrium - " + strconv.Itoa(KyriumBacker), Style: KyriumStyle},
	}

	style := chart.Style{}
	style.FillColor = drawing.Color{131, 135, 142, 255}
	pie := chart.PieChart{
		Width:      512,
		Height:     512,
		Values:     pieValues,
		Background: style,
	}

	filename := "./stats/Backer-PieChart-" + time.Now().Format("Jan 2 15-04-05") + ".png"
	var b bytes.Buffer
	pie.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Backer Distribution ( "+strconv.Itoa(BackerCount)+" registered):", s, m)
	if err != nil {
		return err
	}
	return nil
}

func (h *StatsHandler) NDAPieChart(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	members, err := GetMemberList(s, h.conf)
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
	SponsorBacker := 0
	ContributorBacker := 0
	BackerCount := 0
	//ATVMember := 0

	for _, member := range members {
		for _, roleID := range member.Roles {
			backer := false
			if roleID == h.conf.RolesConfig.ContributorRoleID {
				ContributorBacker = ContributorBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.SponsorRoleID {
				SponsorBacker = SponsorBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.PatronRoleID {
				PatronBacker = PatronBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.GoldRoleID {
				GoldBacker = GoldBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.SapphireRoleID {
				SapphireBacker = SapphireBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.RubyRoleID {
				RubyBacker = RubyBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.EmeraldRoleID {
				EmeraldBacker = EmeraldBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.DiamondRoleID {
				DiamondBacker = DiamondBacker + 1
				backer = true
			}
			if roleID == h.conf.RolesConfig.KyriumRoleID {
				KyriumBacker = KyriumBacker + 1
				backer = true
			}
			if backer {
				BackerCount = BackerCount + 1
			}
		}
	}
	ContributorStyle := chart.Style{FillColor: drawing.Color{R: 252, G: 62, B: 99, A: 255}}
	SponsorStyle := chart.Style{FillColor: drawing.Color{R: 143, G: 143, B: 143, A: 255}}
	PatronStyle := chart.Style{FillColor: drawing.Color{R: 214, G: 187, B: 32, A: 255}}
	GoldStyle := chart.Style{FillColor: drawing.Color{R: 217, G: 176, B: 80, A: 255}}
	SapphireStyle := chart.Style{FillColor: drawing.Color{R: 0, G: 102, B: 255, A: 255}}
	RubyStyle := chart.Style{FillColor: drawing.Color{R: 255, G: 3, B: 3, A: 255}}
	EmeraldStyle := chart.Style{FillColor: drawing.Color{R: 7, G: 230, B: 137, A: 255}}
	DiamondStyle := chart.Style{FillColor: drawing.Color{R: 255, G: 255, B: 255, A: 255}}
	KyriumStyle := chart.Style{FillColor: drawing.Color{R: 219, G: 219, B: 219, A: 255}}

	pieValues := []chart.Value{
		{Value: float64(ContributorBacker), Label: "Patron - " + strconv.Itoa(ContributorBacker), Style: ContributorStyle},
		{Value: float64(SponsorBacker), Label: "Patron - " + strconv.Itoa(SponsorBacker), Style: SponsorStyle},
		{Value: float64(PatronBacker), Label: "Patron - " + strconv.Itoa(PatronBacker), Style: PatronStyle},
		//{Value: float64(ATVMember), Label: "ATV"+strconv.Itoa(PatronBacker)},
		{Value: float64(GoldBacker), Label: "Gold - " + strconv.Itoa(GoldBacker), Style: GoldStyle},
		{Value: float64(SapphireBacker), Label: "Sapphire - " + strconv.Itoa(SapphireBacker), Style: SapphireStyle},
		{Value: float64(RubyBacker), Label: "Ruby - " + strconv.Itoa(RubyBacker), Style: RubyStyle},
		{Value: float64(EmeraldBacker), Label: "Emerald - " + strconv.Itoa(EmeraldBacker), Style: EmeraldStyle},
		{Value: float64(DiamondBacker), Label: "Diamond - " + strconv.Itoa(DiamondBacker), Style: DiamondStyle},
		{Value: float64(KyriumBacker), Label: "Kyrium - " + strconv.Itoa(KyriumBacker), Style: KyriumStyle},
	}

	style := chart.Style{}
	style.FillColor = drawing.Color{131, 135, 142, 255}
	pie := chart.PieChart{
		Width:      512,
		Height:     512,
		Values:     pieValues,
		Background: style,
	}

	filename := "./stats/NDA-PieChart-" + time.Now().Format("Jan 2 15-04-05") + ".png"
	var b bytes.Buffer
	pie.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: NDA Authorized Distribution ( "+strconv.Itoa(BackerCount)+" registered):", s, m)
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

func (h *StatsHandler) ShowInvisibleUsers(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
	if err != nil {
		return err
	}

	users, err := h.GetInvisibleUsers(guild)
	if err != nil {
		return err
	}
	output := ":spy: Current Invisible Users: ```\n"
	for _, user := range users {
		output = output + user.ID + "\n"
	}
	output = output + "\n```\n"
	s.ChannelMessageSend(m.ChannelID, output)
	return nil
}

func (h *StatsHandler) ShowDNDUsers(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	guild, err := s.Guild(h.conf.DiscordConfig.GuildID)
	if err != nil {
		return err
	}

	users, err := h.GetDNDUsers(guild)
	if err != nil {
		return err
	}
	output := ":spy: Current Do Not Disturb Users: ```\n"
	for _, user := range users {
		output = output + user.ID + "\n"
	}
	output = output + "\n```\n"
	s.ChannelMessageSend(m.ChannelID, output)
	return nil
}

func (h *StatsHandler) UserCountChart(days int, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	records, err := h.statsdb.GetFullDB()
	if err != nil {
		return err
	}
	/*combinedRecords, err := h.CombineRecordsByDate(records)
	if err != nil {
		return err
	}*/
	current, err := h.GetCurrentStats(s)
	if err != nil {
		return err
	}
	//dateSplit := strings.Split(current.Date, " ")
	//current.Date = dateSplit[0]
	records = append(records, current)

	dateValues := []time.Time{}
	totalUsers := []float64{}

	for _, record := range records {
		date, err := time.Parse("2006-01-02  15:04:05", record.Date)
		if err != nil {
			return err
		}
		if time.Since(date) < time.Duration(int64(time.Hour)*24*int64(days)) {
			dateValues = append(dateValues, date)
			totalUsers = append(totalUsers, float64(record.TotalUsers))
		}
	}

	style := chart.Style{}
	style.FillColor = drawing.Color{131, 135, 142, 0}
	//style.Show = true
	style.StrokeColor = chart.GetDefaultColor(0).WithAlpha(64)
	style.FillColor = chart.GetDefaultColor(0).WithAlpha(64)
	graph := chart.Chart{
		Title: "Total User Count",
		//Background: chart.StyleShow(),
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateValues,
				YValues: totalUsers,
			},
		},
		XAxis: chart.XAxis{
			Name: "Dates",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name: "Users",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: h.FloatNormalize,
		},
	}

	filename := "./stats/TotalUsersChart-" + time.Now().Format("Jan 2 15-04-05") + ".png"
	var b bytes.Buffer
	graph.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Total User Count - "+strconv.Itoa(records[len(records)-1].TotalUsers)+" users currently registered to this discord in past "+strconv.Itoa(days)+" days:", s, m)

	if err != nil {
		return err
	}
	return nil
}

func (h *StatsHandler) DailyActiveChart(days int, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	records, err := h.statsdb.GetFullDB()
	current, err := h.GetCurrentStats(s)
	if err != nil {
		return err
	}
	records = append(records, current)

	dateValues := []time.Time{}
	totalUsers := []float64{}

	for _, record := range records {
		date, err := time.Parse("2006-01-02 15:04:05", record.Date)
		if err != nil {
			return err
		}
		if record.ActiveUserCount > 0 {
			record.Engagement = record.ActiveUserCount
		}
		if time.Since(date) < time.Duration(int64(time.Hour)*24*int64(days)) {
			dateValues = append(dateValues, date)
			totalUsers = append(totalUsers, float64(record.Engagement))
		}
	}

	style := chart.Style{}
	style.FillColor = drawing.Color{131, 135, 142, 0}
	//style.Show = true
	style.StrokeColor = chart.GetDefaultColor(0).WithAlpha(64)
	style.FillColor = chart.GetDefaultColor(0).WithAlpha(64)
	graph := chart.Chart{
		Title: "Total User Count",
		//Background: chart.StyleShow(),
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateValues,
				YValues: totalUsers,
			},
		},
		XAxis: chart.XAxis{
			Name: "Dates",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name: "Users",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: h.FloatNormalize,
		},
	}

	filename := "./stats/ActiveEngagedUsersChart-" + time.Now().Format("Jan 2 15-04-05") + ".png"
	var b bytes.Buffer
	graph.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Daily Active User Count For Past "+strconv.Itoa(days)+" Days:", s, m)
	if err != nil {
		return err
	}
	return nil
}

func (h *StatsHandler) DailyVoiceChart(days int, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	records, err := h.statsdb.GetFullDB()
	current, err := h.GetCurrentStats(s)
	if err != nil {
		return err
	}
	records = append(records, current)

	dateValues := []time.Time{}
	totalUsers := []float64{}

	for _, record := range records {
		date, err := time.Parse("2006-01-02 15:04:05", record.Date)
		if err != nil {
			return err
		}
		if time.Since(date) < time.Duration(int64(time.Hour)*24*int64(days)) {
			dateValues = append(dateValues, date)
			totalUsers = append(totalUsers, float64(record.VoiceUsers))
		}
	}

	style := chart.Style{}
	style.FillColor = drawing.Color{131, 135, 142, 0}
	//style.Show = true
	style.StrokeColor = chart.GetDefaultColor(0).WithAlpha(64)
	style.FillColor = chart.GetDefaultColor(0).WithAlpha(64)
	graph := chart.Chart{
		Title: "Total User Count",
		//Background: chart.StyleShow(),
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateValues,
				YValues: totalUsers,
			},
		},
		XAxis: chart.XAxis{
			Name: "Dates",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name: "Users",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: h.FloatNormalize,
		},
	}

	filename := "./stats/ActiveEngagedUsersChart-" + time.Now().Format("Jan 2 15-04-05") + ".png"
	var b bytes.Buffer
	graph.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Daily Active Voice User Count For Past "+strconv.Itoa(days)+" Days:", s, m)
	if err != nil {
		return err
	}
	return nil
}

func (h *StatsHandler) DailyOnlineChart(days int, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	records, err := h.statsdb.GetFullDB()
	current, err := h.GetCurrentStats(s)
	if err != nil {
		return err
	}
	records = append(records, current)

	dateValues := []time.Time{}
	totalUsers := []float64{}

	for _, record := range records {
		date, err := time.Parse("2006-01-02 15:04:05", record.Date)
		if err != nil {
			return err
		}
		if time.Since(date) < time.Duration(int64(time.Hour)*24*int64(days)) {
			dateValues = append(dateValues, date)
			totalUsers = append(totalUsers, float64(record.OnlineUsers+record.IdleUsers+record.DNDUsers+record.InvisibleUsers))
		}
	}

	style := chart.Style{}
	style.FillColor = drawing.Color{131, 135, 142, 0}
	//style.Show = true
	style.StrokeColor = chart.GetDefaultColor(0).WithAlpha(64)
	style.FillColor = chart.GetDefaultColor(0).WithAlpha(64)
	graph := chart.Chart{
		Title: "Total User Count",
		//Background: chart.StyleShow(),
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateValues,
				YValues: totalUsers,
			},
		},
		XAxis: chart.XAxis{
			Name: "Dates",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name: "Users",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: h.FloatNormalize,
		},
	}

	filename := "./stats/DailyOnlineUsersChart-" + time.Now().Format("Jan 2 15-04-05") + ".png"
	var b bytes.Buffer
	graph.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Daily Online User Count For Past "+strconv.Itoa(days)+" Days:", s, m)
	if err != nil {
		return err
	}
	return nil
}

func (h *StatsHandler) DailyIdleChart(days int, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {

	records, err := h.statsdb.GetFullDB()
	current, err := h.GetCurrentStats(s)
	if err != nil {
		return err
	}
	records = append(records, current)

	dateValues := []time.Time{}
	totalUsers := []float64{}

	for _, record := range records {
		date, err := time.Parse("2006-01-02 15:04:05", record.Date)
		if err != nil {
			return err
		}
		if time.Since(date) < time.Duration(int64(time.Hour)*24*int64(days)) {
			dateValues = append(dateValues, date)
			totalUsers = append(totalUsers, float64(record.IdleUsers))
		}
	}

	style := chart.Style{}
	style.FillColor = drawing.Color{131, 135, 142, 0}
	//style.Show = true
	style.StrokeColor = chart.GetDefaultColor(0).WithAlpha(64)
	style.FillColor = chart.GetDefaultColor(0).WithAlpha(64)
	graph := chart.Chart{
		Title: "Total User Count",
		//Background: chart.StyleShow(),
		Series: []chart.Series{
			chart.TimeSeries{
				XValues: dateValues,
				YValues: totalUsers,
			},
		},
		XAxis: chart.XAxis{
			Name: "Dates",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: chart.TimeHourValueFormatter,
		},
		YAxis: chart.YAxis{
			Name: "Users",
			//NameStyle:      chart.StyleShow(),
			Style:          style,
			ValueFormatter: h.FloatNormalize,
		},
	}

	filename := "./stats/DailyIdleUsersChart-" + time.Now().Format("Jan 2 15-04-05") + ".png"
	var b bytes.Buffer
	graph.Render(chart.PNG, &b)
	err = h.WriteAndSend(filename, b, ":bar_chart: Daily Idle User Count For Past "+strconv.Itoa(days)+" Days:", s, m)
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
			for i, tmpRecord := range combined {
				if record.Date == tmpRecord.Date {
					found = true
					combined[i].Engagement = tmpRecord.Engagement + record.Engagement
					combined[i].MessageCount = tmpRecord.MessageCount + record.MessageCount
					combined[i].VoiceUsers = tmpRecord.VoiceUsers + record.VoiceUsers
					combined[i].GamingUsers = tmpRecord.GamingUsers + record.GamingUsers
					combined[i].IdleUsers = tmpRecord.IdleUsers + record.IdleUsers
					combined[i].OnlineUsers = tmpRecord.OnlineUsers + record.OnlineUsers
					combined[i].Engagement = tmpRecord.Engagement + record.Engagement + record.ActiveUserCount
					combined[i].TotalUsers = record.TotalUsers
				}
			}
			if !found {
				combined = append(combined, record)
			}
		}
	}
	return combined, nil
}

func (h *StatsHandler) GetUserCount(s *discordgo.Session) (count int, err error) {
	list, err := GetMemberList(s, h.conf)
	if err != nil {
		return 0, err
	}

	return len(list), nil
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

func (h *StatsHandler) GetInvisibleCount(guild *discordgo.Guild) (count int, err error) {
	for _, presence := range guild.Presences {
		if presence.Status == "invisible" {
			count = count + 1
		}
	}
	return count, nil
}

func (h *StatsHandler) GetInvisibleUsers(guild *discordgo.Guild) (users []*discordgo.User, err error) {
	for _, presence := range guild.Presences {
		if presence.Status == "invisible" {
			users = append(users, presence.User)
		}
	}
	return users, nil
}

func (h *StatsHandler) GetDNDUsers(guild *discordgo.Guild) (users []*discordgo.User, err error) {
	for _, presence := range guild.Presences {
		if presence.Status == "dnd" {
			users = append(users, presence.User)
		}
	}
	return users, nil
}

func (h *StatsHandler) GetDNDCount(guild *discordgo.Guild) (count int, err error) {
	for _, presence := range guild.Presences {
		if presence.Status == "dnd" {
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

func (h *StatsHandler) RepairTotals(guild *discordgo.Guild) (err error) {
	records, err := h.statsdb.GetFullDB()
	if err != nil {
		return err
	}

	for i, record := range records {

		if record.TotalUsers == 0 {
			//return errors.New("Found record with null value")
			err = h.statsdb.RemoveStatFromDB(record)
			if err != nil {
				return err
			}
			records[i].TotalUsers = records[i-1].TotalUsers
			err = h.statsdb.AddStatToDB(records[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
