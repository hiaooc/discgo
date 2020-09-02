package slackbot

import (
	"github.com/bl1nk/discord-bot-without-a-fancy-name/datastore"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"strings"
	"time"
)

type slackBot struct {
	dataStore *datastore.DataStore
}

func New(dataStore *datastore.DataStore) *slackBot {
	return &slackBot{dataStore: dataStore}
}

func (s *slackBot) Handler(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	for key, value := range s.dataStore.Responses {
		if !contains(messageCreate.Content, key) {
			continue
		}

		responseMessage := selectRandomEntry(value)

		_, err := session.ChannelMessageSend(messageCreate.ChannelID, responseMessage)

		log.Printf(`Returned message "%s"`, responseMessage)

		if err != nil {
			log.Print(err)
		}

		return
	}

}

func contains(a string, b string) bool {
	return strings.Contains(strings.ToLower(a), strings.ToLower(b))
}

func selectRandomEntry(list []string)  string {
	rand.Seed(time.Now().Unix())

	index := rand.Int() % len(list)
	return list[index]
}