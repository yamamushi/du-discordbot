package main

import (
		"errors"
		"github.com/bwmarrin/discordgo"
			"net/http"
	//"strconv"
	"strings"
	"time"
	"strconv"
	"os"
	"io"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"github.com/lunixbochs/vtclean"
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
	/*
	if command == "estimatesutime" || command == "sutime" || command == "su-convert" {
		if VerifyNDAChannel(m.ChannelID, h.conf){
			if len(payload) < 2 {
				s.ChannelMessageSend(m.ChannelID, command + " expects two arguments: <su> <speed>")
				return
			}
			estimate, err := h.SUToMinutes(payload[0], payload[1])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error: " + err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Estimated travel time: " + estimate)
			return
		}
	}
	*/
	if command == "profilemosaic" {
		if !user.Owner {
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Storing profile images cache")
		// Grab our mosaic images
		err = h.GenerateImageCache(s, m)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error storing images: "+err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Images stored")
		return
	}
	if command == "say" {
		if !user.Moderator {
			return
		}
		if len(payload) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Command say requires two arguments: <channel> <message>")
			return
		}
		channelID := payload[0]

		message := ""
		for i, word := range payload {
			if i > 0 {
				message = message + word + " "
			}
		}
		h.Say(channelID, message, s, m)
		return
	}
}

// UnfoldURL function
func (h *UtilitiesHandler) Say(channelID string, message string, s *discordgo.Session, m *discordgo.MessageCreate) {
	channelID = CleanChannel(channelID)

	if strings.Contains(strings.ToLower(message), "🐰" ) || strings.Contains(strings.ToLower(message), "🐇" ) {
		s.ChannelMessageSend(m.ChannelID, "https://www.tenor.co/zBGa.gif")
		return
	}
	s.ChannelMessageSend(channelID, message)
	s.ChannelMessageSend(m.ChannelID, "Message sent to <#" + channelID + ">")
	return
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
			//fmt.Println(line)
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
	output = output + "Founders pledging has ended, stay tuned for supporter packs in Q1 2018!\n"
	output = output + "Gold and higher Founders can refer to the forums for the Pre-Alpha Testing Schedule. They can also use the ~forumauth command to gain" +
		"access to the NDA channels of this discord.\n"
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
	output = output + "Founders pledging has ended, stay tuned for supporter packs in Q1 2018!"
	output = output + "```\n"

	return output

}

// GetParisTime function
func (h *UtilitiesHandler) GetParisTime() (output string) {

	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return ""
	}

	output = "Current time in Paris is - "
	output = output + time.Now().In(paris).Format("15:04:05 PM")

	return output
}

func (h *UtilitiesHandler) FinalCountdown() (output string) {

	return ":rotating_light: :rotating_light: :rotating_light: !!! https://www.youtube.com/watch?v=9jK-NcRmVcw"

}

func (h *UtilitiesHandler) NovaWrimo() (output string) {

	output = ":rotating_light: :rotating_light: :rotating_light: !!!"
	/*
		output += "\n" + "NovaWrimo 2017 is now live!"
		output += "\n" + "Enter for your chance to win a free Gold Founders Pledge! Find out more at the link below:\n"
		output += "\n" + "https://board.dualthegame.com/index.php?/topic/10415-novawrimo-contest-2016-rules-to-participate/"
	*/
	output += "\n" + "Novawrimo 2017 has completed, stay tuned for another in the future!"

	return output
}

func (h *UtilitiesHandler) Events() (output string) {

	output = ":rotating_light: :rotating_light: :rotating_light: !!! "

	/*
		output += "\n" + "Current Events in Dual Universe!"
		output += "\n" + "```"

		output += "\n" + "It's contest season in Dual Universe! Take part in the following event(s) for your"
		output += "opportunity to win cool stuff!"
		output += "\n" + "NovaWrimo 2017 | https://board.dualthegame.com/index.php?/topic/10415-novawrimo-contest-2016-rules-to-participate/ "
	*/
	output += "\n\n" + "- News -"

	output += "\n" + "As of December 2017, closed pre-alpha testing is currently taking place under a strict NDA policy. "
	output += "If you have a Gold or higher level Founders Pledge, you are currently eligible to take part in these closed tests."
	output += "\n\n" + "Those who would like to discuss testing in this discord can run ~forumauth to authenticate your gold+ account "
	output += "on the forums."

	output += "\n" + "```"

	return output
}

func (h *UtilitiesHandler) SUToMinutes(distance string, speed string) (conversion string, err error){

	distanceFloat, err := strconv.ParseFloat(distance, 64)
	if err != nil {
		return "", err
	}
	if distanceFloat > 100000000 || distanceFloat <= 0 {
		return "", errors.New("Distance value out of bounds")
	}

	speedFloat := 0.0
	if speed == "max" {
		speedFloat = 30000
	} else {
		speedFloat, err = strconv.ParseFloat(speed, 64)
		if err != nil {
			return "", err
		}
	}

	if speedFloat > 100000 || speedFloat <= 0 {
		return "", errors.New("Speed value out of bounds")
	}
	distanceFloat = distanceFloat * 200.00
	secondsInt := (distanceFloat / speedFloat) * 3600.00

	duration := time.Duration(time.Second * time.Duration(secondsInt))
	return duration.String(), nil
}

// GenerateImageCache function
func (h * UtilitiesHandler) GenerateImageCache(s *discordgo.Session, m *discordgo.MessageCreate) (err error){

	// Create directory if not exists
	profile_pics_dir := "./profile_pics"
	err = CreateDirIfNotExist(profile_pics_dir)
	if err != nil {
		return err
	}

	guild, err := s.Guild(s.State.Guilds[0].ID)
	for i, member := range guild.Members{
		photourl := member.User.AvatarURL("")
		response, err := http.Get(photourl)
		if err != nil {
			return err
		}

		defer response.Body.Close()

		if response.ContentLength > 0 {

			filetype := strings.Split(photourl, ".")

			username := strings.Replace(member.User.Username, "/", "_", -1)
			picpath := profile_pics_dir+"/"+username+"."+filetype[len(filetype)-1]

			if _, err := os.Stat(picpath); os.IsNotExist(err) {
				if i != 0 {
					if 50 % i == 0 {
						time.Sleep(time.Duration(time.Second*10))
					}
				}
				//open a file for writing
				file, err := os.Create(picpath)
				if err != nil {
					return err
				}
				// Use io.Copy to just dump the response body to the file. This supports huge files
				_, err = io.Copy(file, response.Body)
				if err != nil {
					return err
				}
				file.Close()
			}
		}
	}

	return nil
}