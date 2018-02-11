package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lunixbochs/vtclean"
	"io/ioutil"
	"net/http"
	//"strconv"
	"strings"
	"time"
)

// UtilitiesHandler struct
type UtilitiesHandler struct {
	user     *UserHandler
	db       *DBHandler
	conf     *Config
	registry *CommandRegistry
	logchan  chan string
	callback *CallbackHandler
}

// ShortURLResponse struct
type ShortURLResponse struct {
	Short map[string]string `json:"/short"`
}

// Read function
func (h *UtilitiesHandler) Read(s *discordgo.Session, m *discordgo.MessageCreate) {

	if !SafeInput(s, m, h.conf) {
		return
	}

	command, payload := CleanCommand(m.Content, h.conf)

	h.user.CheckUser(m.Author.ID)

	user, err := h.db.GetUser(m.Author.ID)
	if err != nil {
		//fmt.Println("Error finding user")
		return
	}

	if !user.Citizen {
		return
	}
	/*
		command = payload[0]
		payload = RemoveStringFromSlice(payload, command)
	*/

	if command == "unfold" || command == "unshorten" {
		if len(payload) < 1 {
			s.ChannelMessageSend(m.ChannelID, "<unfold> expects an argument")
			return
		}
		response, err := h.UnfoldURL(payload[0])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not unfold "+payload[0]+" : "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, payload[0]+" unfolds to ```\n"+response+"\n```")
		return
	}
	if command == "moon" {
		message, err := h.GetMoon()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not get moon: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Current Moon From Earth:\n"+message)
		return

	}
	if command == "countdown" {
		s.ChannelMessageSend(m.ChannelID, h.GetCountdownStatus())
		return
	}
	if command == "pledging" || command == "pledges" || command == "crowdfunding" || command == "founderspacks" {

		s.ChannelMessageSend(m.ChannelID, h.GetPledgingStatus())
		return
	}
	if command == "nqtime" || command == "paristime" {
		s.ChannelMessageSend(m.ChannelID, h.GetParisTime())
		return
	}
	if command == "finalcountdown" {
		s.ChannelMessageSend(m.ChannelID, h.FinalCountdown())
		return
	}
	if command == "novawrimo" {
		s.ChannelMessageSend(m.ChannelID, h.NovaWrimo())
		return
	}
	if command == "events" {
		s.ChannelMessageSend(m.ChannelID, h.Events())
		return
	}

}

// UnfoldURL function
func (h *UtilitiesHandler) UnfoldURL(input string) (output string, err error) {

	// The first step is to use golang’s http module to get the response:
	res, err := http.Get("http://x.datasig.io/short?url=" + input)
	if err != nil {
		return output, err
	}

	// Assuming you didnt see a panic call, the response to this http call is
	// being stored in the res variable. Next, we need to read the http body
	// into a byte array for parsing/processing (using golangs ioutil library):

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return output, err
	}

	// Unmarshal our response to a json type
	var unmarshalledresponse = new(ShortURLResponse)
	err = json.Unmarshal(body, &unmarshalledresponse)
	if err != nil {
		return output, err
	}

	if unmarshalledresponse.Short["destination"] == input {
		return output, errors.New("Could not unshorten")
	}

	output = unmarshalledresponse.Short["destination"]
	fmt.Println()
	return output, nil

}

// GetMoon function
func (h *UtilitiesHandler) GetMoon() (output string, err error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://wttr.in/Moon", nil)
	if err != nil {
		return output, err
	}

	req.Header.Set("User-Agent", "curl/1.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return output, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return output, err
	}

	output = string(body)

	output = vtclean.Clean(output, false)
	payload := strings.Split(output, "\n")

	output = "\n```\n"
	for i, line := range payload {
		if i < 25 {
			fmt.Println(line)
			output = output + line + "\n"
		}
	}
	output = output + "\n```"
	return output, nil

}

// GetCountdownStatus function returns a formatted list of days important to Alpha and pledging
func (h *UtilitiesHandler) GetCountdownStatus() (output string) {

	////atvalpha := time.Date(2017, 9, 23, 13, 0, 0, 0, time.UTC)
	//endofpledges := time.Date(2017, 9, 8, 8, 0, 0, 0, time.UTC)

	//exacttimeuntilprealpha := endofsep.Sub(time.Now().UTC())
	//exacttimeuntilatvalpha := atvalpha.Sub(time.Now().UTC())
	//exacttimeuntilpledges := endofpledges.Sub(time.Now().UTC())

	//daysuntilprealpha := endofsep.YearDay() - time.Now().YearDay()
	//daysuntilatvalpha := atvalpha.YearDay() - time.Now().YearDay()
	//daysuntilpledges := endofpledges.YearDay() - time.Now().YearDay()

	output = "Current Important Countdowns: ```\n"
	//output = output + "Time Until Founders Pack Pledging Ends: " + strconv.Itoa(daysuntilpledges) + " days (Approx: " + TruncateTime(exacttimeuntilpledges, time.Second).String() + ")\n"
	output = output + "Founders pledging has ended, stay tuned for supporter packs in Q4 2017!\n"
	output = output + "Gold and higher Founders can refer to the forums for the Pre-Alpha Testing Schedule.\n"
	//output = output + "Time Until ATV Pre-Alpha Release      : " + strconv.Itoa(daysuntilatvalpha) + " days (Approx: " + TruncateTime(exacttimeuntilatvalpha, time.Second).String() + ")\n"
	//output = output + "Time Until ATV Pre-Alpha Release      : ATV in Pre-Alpha, Have Fun!\n"
	//output = output + "Time Until Pre-Alpha Release          : " + strconv.Itoa(daysuntilprealpha) + " days (Approx: " + TruncateTime(exacttimeuntilprealpha, time.Second).String() + ")\n"
	//output = output + "Time Until Pre-Alpha Release          : Gold+ Backers in Pre-Alpha, Have Fun!\n"
	output = output + "```\n"

	return output

}

// GetPledgingStatus function
func (h *UtilitiesHandler) GetPledgingStatus() (output string) {

	//endofpledges := time.Date(2017, 9, 8, 8, 0, 0, 0, time.UTC)
	//exacttimeuntilpledges := endofpledges.Sub(time.Now().UTC())
	//daysuntilpledges := endofpledges.YearDay() - time.Now().YearDay()

	output = "Current Pledging Information: ```\n"
	//output = output + "Time Until Founders Pack Pledging Ends: " + strconv.Itoa(daysuntilpledges) + " days (Exactly: " + TruncateTime(exacttimeuntilpledges, time.Second).String() + ")\n"
	output = output + "Founders pledging has ended, stay tuned for supporter packs in Q4 2017!"
	output = output + "```\n"

	return output

}

// GetParisTime function
func (h *UtilitiesHandler) GetParisTime() (output string){

	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return ""
	}

	output = "Current time in Paris is - "
	output = output + time.Now().In(paris).Format("15:04:05 PM")

	return output
}


func (h *UtilitiesHandler) FinalCountdown() (output string){

	return ":rotating_light: :rotating_light: :rotating_light: !!! https://www.youtube.com/watch?v=9jK-NcRmVcw"

}

func (h *UtilitiesHandler) NovaWrimo() (output string){

	output = ":rotating_light: :rotating_light: :rotating_light: !!!"
	output += "\n" + "NovaWrimo 2017 is now live!"
	output += "\n" + "Enter for your chance to win a free Gold Founders Pledge! Find out more at the link below:\n"
	output += "\n" + "https://board.dualthegame.com/index.php?/topic/10415-novawrimo-contest-2016-rules-to-participate/"
	return output
}

func (h *UtilitiesHandler) Events() (output string){

	output = ":rotating_light: :rotating_light: :rotating_light: !!! "
	output += "\n" + "Current Events in Dual Universe!"
	output += "\n" + "```"

	output += "\n" + "It's contest season in Dual Universe! Take part in the following event(s) for your"
	output += "opportunity to win cool stuff!"
	output += "\n" + "NovaWrimo 2017 | https://board.dualthegame.com/index.php?/topic/10415-novawrimo-contest-2016-rules-to-participate/ "

	output += "\n\n" + "- News -"

	output += "\n" + "As of December 2017, closed pre-alpha testing is currently taking place under a strict NDA policy. "
	output += "If you have a Gold or higher level Founders Pledge, you are currently eligible to take part in these closed tests."
	output += "\n\n" + "Those who would like to discuss testing in this discord can run ~forumauth to authenticate your gold+ account "
	output +=  "on the forums."


	output += "\n" + "```"


	return output
}