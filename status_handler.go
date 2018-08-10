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

	h.FormatSendStatus(status, s, m)
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
	fmt.Println("Tables: " + strconv.Itoa(len(statusTableBox)))

	if len(statusTableBox) > 0 {

		statusTableRows := statusTableBox[0].Find("table", "class", "table").FindAll("tr")

		fmt.Println("Rows: " + strconv.Itoa(len(statusTableRows)))

		if len(statusTableRows) > 1 {

			for rowNum, row := range statusTableRows {
				// We always skip the first row, which is a description row
				var status ServerStatus
				if rowNum > 0 {
					rowColumns := row.FindAll("td")
					fmt.Println("Row "+strconv.Itoa(rowNum)+ " Column Count: " + strconv.Itoa(len(rowColumns)))

					for colNumber, column := range rowColumns {
						fmt.Println(strings.TrimSpace(column.Text()))
						if colNumber == 0 {
							status.Status = strings.TrimSpace(column.Text())
						}
						if colNumber == 1 {
							status.TestType = strings.TrimSpace(column.Text())
						}
						if colNumber == 2 {
							status.Access = strings.TrimSpace(column.Text())
						}
						if colNumber == 3 {
							status.StartDate, err  = time.Parse("January 02, 2006 03:04 PM MST", strings.TrimSpace(column.Text()))
							if err != nil {
								return statuslist, errors.New("Could not parse start date for row " + strconv.Itoa(rowNum))
							}
						}
						if colNumber == 4 {
							status.EndDate, err  = time.Parse("January 02, 2006 03:04 PM MST", strings.TrimSpace(column.Text()))
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
					fmt.Println("\n")
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