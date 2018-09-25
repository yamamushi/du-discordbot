package main

import (
	"github.com/bwmarrin/discordgo"
	//"time"
	"time"
	"fmt"
)

type LanderHandler struct {
	configdb *ConfigDB
	user *UserHandler
}


// Read function
// This handler will wait when a new user joins, and automatically assign roles to users who have not yet authenticated properly
// After a five minute period.
func (h *LanderHandler) Read(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	unlandedRoleID, err := getRoleIDByName(s, s.State.Guilds[0].ID, "Unlanded")
	if err != nil {
		return
	}
	s.GuildMemberRoleAdd(s.State.Guilds[0].ID, m.User.ID, unlandedRoleID)

	landingZoneID, err := getChannelIDByName(s, s.State.Guilds[0].ID, "landing_pad")
	serverinfoID, err := getChannelIDByName(s, s.State.Guilds[0].ID, "server-information")

	s.ChannelMessageSend(landingZoneID, "Welcome to the Official **Dual Universe** Discord Server "+
		m.User.Mention()+ "! Please take a moment to read <#"+ serverinfoID+ "> to find your way in to the rest of the server."+
		" (This is just the lobby, there are other channels here)")

	h.user.CheckUser(m.User.ID)

	autoland, err := h.configdb.CheckEnabled("autoland")
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	if autoland {
		landedRoleID, err := getRoleIDByName(s, s.State.Guilds[0].ID, "Landed")
		if err != nil {
			return
		}

		spectatorRoleID, err := getRoleIDByName(s, s.State.Guilds[0].ID, "Guest")
		if err != nil {
			return
		}

		time.Sleep(time.Duration(time.Minute * 2))
		s.GuildMemberRoleRemove(s.State.Guilds[0].ID, m.User.ID, unlandedRoleID)

		time.Sleep(time.Duration(time.Second * 1))
		err = s.GuildMemberRoleAdd(s.State.Guilds[0].ID, m.User.ID, landedRoleID)
		if err == nil {
			s.ChannelMessageSend(landingZoneID, m.User.Username+" has been added to the Landed role")
		}

		time.Sleep(time.Duration(time.Second * 1))
		s.GuildMemberRoleAdd(s.State.Guilds[0].ID, m.User.ID, spectatorRoleID)

		time.Sleep(time.Duration(time.Second * 1))
		newcomersChannelID, err := getChannelIDByName(s, s.State.Guilds[0].ID, "newcomers")
		s.ChannelMessageSend(newcomersChannelID, m.User.Mention()+" has landed")
	}
	member, err := s.GuildMember(s.State.Guilds[0].ID, m.User.ID)
	if err != nil {
		return
	}
	s.State.MemberAdd(member)

	return
}

func (h *LanderHandler) ReadRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	//fmt.Println("Remove event caught")
	// Don't use this as it will malloc panic
	//member, _ := s.GuildMember(s.State.Guilds[0].ID, m.User.ID)
	//s.State.MemberRemove(member)


	// Wipe our autoroles so that when the user rejoins they start at Guest again
	userrecord, _ := h.user.GetUser(m.User.ID)
	userrecord.LatestRoleTimeout = 0
	//userrecord.RabbitCount = 0
	userrecord.HistoricalAutoRoles = []string{""}

	h.user.UpdateUserRecord(userrecord)

	//generalChannelID, _ := getChannelIDByName(s, s.State.Guilds[0].ID, "general")
	//s.ChannelMessageSend(generalChannelID, "<@"+m.User.ID+"> has left the server.")
	return
}

