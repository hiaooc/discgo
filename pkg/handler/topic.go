package handler

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ChangeTopic(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	cmdPrefix := regexp.MustCompile(fmt.Sprintf("^<@!?%s> topic", s.State.User.ID))
	if cmdPrefix.MatchString(m.Content) {
		newTopic := strings.TrimSpace(cmdPrefix.ReplaceAllLiteralString(m.Content, ""))
		log.Printf("setting channel topic for channel %s: %s", m.ChannelID, newTopic)
		_, err := s.ChannelEditComplex(m.ChannelID, &discordgo.ChannelEdit{
			Topic: newTopic,
		})
		reaction := "✅"
		if err != nil {
			reaction = "🚫"
			log.Printf("edit channel: %v\n", err)
		}
		err = s.MessageReactionAdd(m.ChannelID, m.Message.ID, reaction)
		if err != nil {
			log.Printf("add reaction: %v\n", err)
		}
	}
}
