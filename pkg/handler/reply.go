package handler

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hiaooc/discgo/pkg/datastore"
)

type Replier struct {
	ds             *datastore.DataStore
	waitingForUser map[string]string
}

func NewReplier(ds *datastore.DataStore) *Replier {
	return &Replier{
		ds:             ds,
		waitingForUser: map[string]string{},
	}
}

func (r *Replier) Reply(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	for k, v := range r.ds.Contents.Responses {
		if strings.Contains(m.Content, k) {
			_, err := s.ChannelMessageSend(m.ChannelID, selectReply(v))
			if err != nil {
				log.Printf("send message: %v\n", err)
				return
			}
			return
		}
	}
}

func selectReply(replies []string) string {
	return replies[rand.Int()%len(replies)]
}

func (r *Replier) AddReply(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	for userID, trigger := range r.waitingForUser {
		if userID == m.Author.ID {
			delete(r.waitingForUser, userID)
			replies := strings.Split(m.Content, "\n")
			r.ds.Contents.Responses[trigger] = replies

			err := r.ds.Save()
			if err != nil {
				log.Printf("save datastore: %v\n", err)
				return
			}

			err = s.MessageReactionAdd(m.ChannelID, m.Message.ID, "✅")
			if err != nil {
				log.Printf("add reaction: %v\n", err)
				return
			}

			return
		}
	}

	re := regexp.MustCompile(fmt.Sprintf(`^<@!?%s> reply`, s.State.User.ID))
	if !re.MatchString(m.Content) {
		return
	}

	trigger := strings.TrimSpace(re.ReplaceAllLiteralString(m.Content, ""))
	r.waitingForUser[m.Author.ID] = trigger

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(`%s Sure! Reply with responses, one per line`, m.Author.Mention()))
	if err != nil {
		log.Printf("send message: %v\n", err)
		return
	}
}

func (r *Replier) RemoveReply(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	re := regexp.MustCompile(fmt.Sprintf(`^<@!?%s> remove`, s.State.User.ID))
	if !re.MatchString(m.Content) {
		return
	}

	trigger := strings.TrimSpace(re.ReplaceAllLiteralString(m.Content, ""))
	delete(r.ds.Contents.Responses, trigger)
	err := r.ds.Save()
	if err != nil {
		log.Printf("save datastore: %v\n", err)
		return
	}

}
