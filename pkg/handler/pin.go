package handler

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func PinMessage(ds *discordgo.Session, event *discordgo.MessageReactionAdd) {
	if event.UserID == ds.State.User.ID || event.MessageReaction.Emoji.Name != "📌" {
		return
	}

	msg, err := ds.ChannelMessage(event.ChannelID, event.MessageID)
	if err != nil {
		log.Printf("get message state: %v\n", err)
		return
	}
	if msg.Pinned {
		return
	}

	if err := ds.ChannelMessagePin(event.ChannelID, event.MessageID); err != nil {
		log.Printf("pin message: %v\n", err)
		return
	}
}

func UnpinMessage(ds *discordgo.Session, event *discordgo.MessageReactionRemove) {
	if event.UserID == ds.State.User.ID || event.MessageReaction.Emoji.Name != "📌" {
		return
	}

	msg, err := ds.ChannelMessage(event.ChannelID, event.MessageID)
	if err != nil {
		log.Printf("get message state: %v\n", err)
		return
	}
	for _, r := range msg.Reactions {
		if r.Emoji.Name == "📌" {
			return
		}
	}

	if err := ds.ChannelMessageUnpin(event.ChannelID, event.MessageID); err != nil {
		log.Printf("unpin message: %v", err)
		return
	}
}
