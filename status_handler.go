package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"time"
	"github.com/anaskhan96/soup"
	"strconv"
	"errors"
)

// RecruitmentHandler struct
type ServerStatusHandler struct {
	conf     *Config
	registry *CommandRegistry
	db       *DBHandler
	userdb   *UserHandler
}

type ServerStatus struct {
	Status      string
	StatusColor int
	TestType    string
	Access      string
	StartDate   time.Time
	EndDate     time.Time
	Duration    time.Duration
}


// Init function
func (h *ServerStatusHandler) Init() {
	h.RegisterCommands()
}


// RegisterCommands function
func (h *ServerStatusHandler) RegisterCommands() (err error) {
	h.registry.Register("status", "Display current DU server status", "status")
	h.registry.Register("countdown", "Display upcoming DU server uptime", "countdown")
	return nil
}

// Read function
func (h *ServerStatusHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	cp := h.conf.DUBotConfig.CP

	if !SafeInput(s, m, h.conf) {
		return
	}

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if strings.HasPrefix(m.Content, cp+"status") || strings.HasPrefix(m.Content, cp+"countdown") {
		if h.registry.CheckPermission("status", m.ChannelID, user) || h.registry.CheckPermission("countdown", m.ChannelID, user) {

			command := strings.Fields(m.Content)

			// Grab our sender ID to verify if this user has permission to use this command
			db := h.db.rawdb.From("Users")
			var user User
			err := db.One("ID", m.Author.ID, &user)
			if err != nil {
				fmt.Println("error retrieving user:" + m.Author.ID)
			}

			if user.Citizen {
				h.ParseCommand(command, s, m)
			}
		}
	}
}



// ParseCommand function
func (h *ServerStatusHandler) ParseCommand(commandlist []string, s *discordgo.Session, m *discordgo.MessageCreate) {

	//command, payload := SplitPayload(commandlist)

	statuslist, err := h.GetServerStatusList()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	status, err := h.FindCurrentTest(statuslist)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	h.FormatStatusEmbed(status, s, m)
	return
}


func (h *ServerStatusHandler) GetServerStatusList() (statuslist []ServerStatus, err error) {

	resp, err := soup.Get("https://www.dualthegame.com/en/server-status/") // Append page=1000 so we get the last page
	if err != nil {
		//fmt.Println("Could not retreive page: " + record.ForumProfile)
		return statuslist, err
	}

	doc := soup.HTMLParse(resp)

	statusTableBox := doc.FindAll("div", "class", "table-responsive")
	//fmt.Println("Tables: " + strconv.Itoa(len(statusTableBox)))

	if len(statusTableBox) > 0 {

		statusTableRows := statusTableBox[0].Find("table", "class", "table").FindAll("tr")

		//fmt.Println("Rows: " + strconv.Itoa(len(statusTableRows)))

		if len(statusTableRows) > 1 {

			for rowNum, row := range statusTableRows {
				// We always skip the first row, which is a description row
				var status ServerStatus
				if rowNum > 0 {
					rowColumns := row.FindAll("td")
					//fmt.Println("Row "+strconv.Itoa(rowNum)+ " Column Count: " + strconv.Itoa(len(rowColumns)))

					for colNumber, column := range rowColumns {
						//fmt.Println(strings.TrimSpace(column.Text()))
						if colNumber == 0 {
							classAttrs := column.Attrs()["class"]
							colorString := strings.Trim(strings.Split(classAttrs, " ")[0], "is-")

							if colorString == "green" {
								status.StatusColor = 6932560
							}
							if colorString == "gray" {
								status.StatusColor = 14013909
							}
							if colorString == "yellow" {
								status.StatusColor = 16380271
							}
							if colorString == "orange" {
								status.StatusColor = 16743941
							}
							if colorString == "red" {
								status.StatusColor = 14417920
							}
							status.Status = strings.TrimSpace(column.Text())
						}
						if colNumber == 1 {
							status.TestType = strings.TrimSpace(column.Text())
						}
						if colNumber == 2 {
							if len(column.Text()) > 0 {
								status.Access = strings.TrimSpace(column.Text())
							} else {
								afield := column.Find("a")
								status.Access = afield.Text()
							}
						}
						if colNumber == 3 {
							//fmt.Println(strings.TrimSpace(column.Text()))
							status.StartDate, err  = time.Parse("January 02, 2006 - 15:04 MST", strings.TrimSpace(column.Text()))
							if err != nil {
								//fmt.Println(err.Error())
								return statuslist, errors.New("Could not parse start date for row " + strconv.Itoa(rowNum))
							}
						}
						if colNumber == 4 {
							status.EndDate, err  = time.Parse("January 02, 2006 - 15:04 MST", strings.TrimSpace(column.Text()))
							if err != nil {
								return statuslist, errors.New("Could not parse end date for row " + strconv.Itoa(rowNum))
							}
						}
						if colNumber == 5 {
							timeFields := strings.Split(strings.TrimSpace(column.Text()), ":")
							if len(timeFields) < 2 {
								return statuslist, errors.New("Could not parse duration for row " + strconv.Itoa(rowNum))
							}
							hoursparsed, err := strconv.Atoi(timeFields[0])
							if err != nil {
								return statuslist, errors.New("Could not parse hour duration for row " + strconv.Itoa(rowNum) + " - " + err.Error())
							}
							minutesparsed, err := strconv.Atoi(timeFields[1])
							if err != nil {
								return statuslist, errors.New("Could not parse minute duration for row " + strconv.Itoa(rowNum) + " - " + err.Error())
							}

							hours := time.Duration(time.Hour * time.Duration(hoursparsed))
							minutes := time.Duration(time.Minute * time.Duration(minutesparsed))
							duration := time.Duration(hours + minutes)
							status.Duration = duration
						}

					}
					statuslist = append(statuslist, status)
					//fmt.Println("\n")
				}
			}

		} else {
			return statuslist, errors.New("Could not parse status table correctly")
		}

	}

	return statuslist, nil
}


func (h *ServerStatusHandler) FindCurrentTest(statuslist []ServerStatus) (status ServerStatus, err error) {

	for num, item := range statuslist {
		if strings.ToLower(item.Status) == "live" {
			return item, nil
		}

		if time.Now().Before(item.StartDate) {
			if num == 0 {
				return item, nil
			}
			if time.Now().After(statuslist[num-1].EndDate) {
				return item, nil
			}
		}
	}

	return status, errors.New("Could not find current or next test")
}


func (h *ServerStatusHandler) FormatSendStatus(status ServerStatus,  s *discordgo.Session, m *discordgo.MessageCreate) {

	output := ":satellite: Server Status \n```\n"

	loc, _ := time.LoadLocation("America/Chicago")
	if strings.ToLower(status.Status) == "live" {

		output += "The server is currently live until:\n" +
			status.EndDate.In(loc).Format("January 02, 2006 03:04 PM MST") +
			" ("+status.EndDate.Format("January 02, 2006 03:04 PM MST")  + ")\n\n"

		output += "The current "+status.TestType+" test session is scheduled to end in approximately:\n" +
			fmtDuration(status.EndDate.Sub(time.Now().Round(0))) + "\n"

	} else {

		output += "The server is currently offline, and is scheduled to come online at:\n " +
			status.StartDate.In(loc).Format("January 02, 2006 03:04 PM MST")+
			" ("+status.EndDate.Format("January 02, 2006 03:04 PM MST")  + ")\n\n"

		output += "The next "+status.TestType+" test session is scheduled to begin in approximately:\n" +
			status.StartDate.Sub(time.Now().Round(0)).String()

		output += "The current scheduled upcoming test duration is " +
			fmtDuration(status.Duration)
	}

	output += "\n```\n"

	s.ChannelMessageSend(m.ChannelID, output )
	return
}

func (h *ServerStatusHandler) FormatStatusEmbed(status ServerStatus,  s *discordgo.Session, m *discordgo.MessageCreate) {

	output := discordgo.MessageEmbed{}
	authorField := discordgo.MessageEmbedAuthor{}
	output.Author = &authorField
	output.Color = status.StatusColor
	authorField.Name = "Dual Universe " + status.TestType + " Test"


	fields := []*discordgo.MessageEmbedField{}
	//counter := discordgo.MessageEmbedField{}

	loc, _ := time.LoadLocation("America/Chicago")

	//output.Timestamp = time.Now().In(loc).Format("January 02 03:04 PM MST")
	if strings.ToLower(status.Status) == "live" {

		output.Title = "The server is currently **Live**"

		output.Footer = &discordgo.MessageEmbedFooter{Text:"Test Ending In: " + fmtDuration(status.EndDate.Sub(time.Now().Round(0))),
			IconURL:"https://cdn.discordapp.com/attachments/418457755276410880/473080359219625989/Server_Logo.jpg"}
		//counter.Name = "Ending In"
		//counter.Value = fmtDuration(status.EndDate.Sub(time.Now().Round(0)))

		//output += "The current "+status.TestType+" test session is scheduled to end in approximately:\n" +
		//	fmtDuration(status.EndDate.Sub(time.Now().Round(0))) + "\n"

	} else if strings.ToLower(status.Status) == "planned"{

		output.Title = "The server is currently scheduled to come online."

		output.Footer = &discordgo.MessageEmbedFooter{Text:"Test Starting In: " + fmtDuration(status.StartDate.Sub(time.Now().Round(0))),
			IconURL:"https://cdn.discordapp.com/attachments/418457755276410880/473080359219625989/Server_Logo.jpg"}
		//counter.Name = "Starting In"
		//counter.Value = fmtDuration(status.StartDate.Sub(time.Now().Round(0)))

		//output += "The next "+status.TestType+" test session is scheduled to begin in approximately:\n" +
		//	status.StartDate.Sub(time.Now().Round(0)).String()

		//output += "The current scheduled upcoming test duration is " +
		//	fmtDuration(status.Duration)
	}

	output.URL = "https://www.dualthegame.com/en/server-status/"

	output.Thumbnail = &discordgo.MessageEmbedThumbnail{URL:"https://cdn.discordapp.com/attachments/452882025553199105/477363590429409283/ux5Nv-CRwNwu9nedK_6cr4HfMqTeeC65Hz3Rnxz6FNg.png"}

	//output.Image = &discordgo.MessageEmbedImage{URL:"https://cdn.discordapp.com/attachments/327676701133897730/477360048494870538/serverstatus.png", Width:50, Height:50}

	//fields = append(fields, &counter)

	access := discordgo.MessageEmbedField{}
	access.Name = "Access"
	access.Value = strings.TrimPrefix(status.Access, "Access for ")
	access.Value = strings.TrimSpace(access.Value)
	if access.Value == "Pledge now and get access" {
		access.Value = "[Pledge now and get access to the next test!](https://www.dualthegame.com/en/pledge)"
	}
	fields = append(fields, &access)


	duration := discordgo.MessageEmbedField{}
	duration.Name = "Duration"
	duration.Value = fmtDuration(status.Duration)
	fields = append(fields, &duration)


	startTime := discordgo.MessageEmbedField{}
	startTime.Inline = true
	startTime.Name = "Start Date"
	startTime.Value = status.StartDate.In(loc).Format("January 02 03:04 PM MST")+
		"\n"+status.StartDate.Format("January 02 03:04 PM MST")
	fields = append(fields, &startTime)


	endTime := discordgo.MessageEmbedField{}
	endTime.Inline = true
	endTime.Name = "End Date"
	endTime.Value = status.EndDate.In(loc).Format("January 02 03:04 PM MST") +
		"\n"+status.EndDate.Format("January 02 03:04 PM MST")
	fields = append(fields, &endTime)



	output.Fields = fields


	//output += "\n```\n"

	sentEmbed, err := s.ChannelMessageSendEmbed(m.ChannelID, &output )
	if err != nil {
		return
	}

	if m.ChannelID != "477247944249049089" {
		time.Sleep(time.Second*30)
		err = s.ChannelMessageDelete(m.ChannelID, sentEmbed.ID)
		if err != nil {
			return
		}
		err = s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			return
		}
	}
	return
}