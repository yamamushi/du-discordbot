package main

import "github.com/bwmarrin/discordgo"

func (h *InfoHandler) RenderSatellitePage(record InfoRecord, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	embed := &discordgo.MessageEmbed{}
	embed.Title = record.Name

	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return nil
}
