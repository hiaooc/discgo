package handler

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func reactWithCheckmark(s *discordgo.Session, channelID, messageID string){
	err := s.MessageReactionAdd(channelID, messageID, "✅")
	if err != nil {
		log.Printf("add reaction: %v\n", err)
		return
	}
}
