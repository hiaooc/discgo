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

func (r *Replier) Handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	for userID, trigger := range r.waitingForUser {
		if userID == m.Author.ID {
			r.addReply(s, m, userID, trigger)
			return
		}
	}

	removeReplyTrigger := regexp.MustCompile(fmt.Sprintf(`^<@!?%s> remove`, s.State.User.ID))
	if removeReplyTrigger.MatchString(m.Content) {
		trigger := strings.TrimSpace(removeReplyTrigger.ReplaceAllLiteralString(m.Content, ""))
		r.removeReply(s, m, trigger)
		return
	}

	addReplyTrigger := regexp.MustCompile(fmt.Sprintf(`^<@!?%s> reply`, s.State.User.ID))
	if addReplyTrigger.MatchString(m.Content) {
		trigger := strings.TrimSpace(addReplyTrigger.ReplaceAllLiteralString(m.Content, ""))
		r.addReplyTrigger(s, m, trigger)
		return
	}

	r.reply(s, m)
}

func (r *Replier) addReply(s *discordgo.Session, m *discordgo.MessageCreate, userID, trigger string) {
	delete(r.waitingForUser, userID)
	replies := strings.Split(m.Content, "\n")
	r.ds.Contents.Responses[trigger] = replies

	err := r.ds.Save()
	if err != nil {
		log.Printf("save datastore: %v\n", err)
		return
	}

	reactWithCheckmark(s, m.ChannelID, m.Message.ID)
}

func (r *Replier) reply(s *discordgo.Session, m *discordgo.MessageCreate) {
	for k, v := range r.ds.Contents.Responses {
		if strings.Contains(m.Content, k) {
			_, err := s.ChannelMessageSend(m.ChannelID, v[rand.Int()%len(v)])
			if err != nil {
				log.Printf("send message: %v\n", err)
				return
			}
			return
		}
	}
}

func (r *Replier) addReplyTrigger(s *discordgo.Session, m *discordgo.MessageCreate, trigger string) {
	r.waitingForUser[m.Author.ID] = trigger
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(`%s Sure! Reply with responses, one per line`, m.Author.Mention()))
	if err != nil {
		log.Printf("send message: %v\n", err)
		return
	}
}

func (r *Replier) removeReply(s *discordgo.Session, m *discordgo.MessageCreate, trigger string) {
	delete(r.ds.Contents.Responses, trigger)
	err := r.ds.Save()
	if err != nil {
		log.Printf("save datastore: %v\n", err)
		return
	}

	reactWithCheckmark(s, m.ChannelID, m.Message.ID)
}
