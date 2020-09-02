package slackbot

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bl1nk/discgo/datastore"
	"github.com/bwmarrin/discordgo"
)

type slackBot struct {
	dataStore *datastore.DataStore
}

func New(dataStore *datastore.DataStore) *slackBot {
	return &slackBot{dataStore: dataStore}
}

func (s *slackBot) Handler(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	for key, value := range s.dataStore.Contents.Responses {
		if !contains(messageCreate.Content, key) {
			continue
		}

		responseMessage := selectRandomEntry(value)

		_, err := session.ChannelMessageSend(messageCreate.ChannelID, responseMessage)
		if err != nil {
			log.Print(err)
		}

		log.Printf(`trigger: "%s" response: "%s"`, key, responseMessage)
		return
	}
}

func contains(a string, b string) bool {
	return strings.Contains(strings.ToLower(a), strings.ToLower(b))
}

func selectRandomEntry(list []string) string {
	rand.Seed(time.Now().Unix())

	index := rand.Int() % len(list)
	return list[index]
}
