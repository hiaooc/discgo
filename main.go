package main

import (
	"flag"
	"fmt"
	"log"
	ghttp "net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bl1nk/discgo/datastore"
	"github.com/bl1nk/discgo/http"
	"github.com/bl1nk/discgo/slackbot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
)

var (
	dataStorePath = flag.String("datastore", "", "Path to JSON file")
	addr          = flag.String("listen", ":8080", "Listen address")
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

func main() {
	flag.Parse()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Token needs to be configured via env variable BOT_TOKEN")
	}
	if *dataStorePath == "" {
		log.Fatal("-datastore must not be empty")
	}

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	ds, err := datastore.Read(*dataStorePath)
	if err != nil {
		log.Fatal(err)
	}

	bot := slackbot.New(ds)
	discord.AddHandler(bot.Handler)
	discord.AddHandler(ChangeTopic)

	discord.AddHandler(PinMessage)
	discord.AddHandler(UnpinMessage)

	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}

	handler := http.NewHandler(ds)
	router := chi.NewRouter()
	router.Get("/config", handler.ReadConfig)
	router.Post("/config", handler.WriteConfig)
	srv := ghttp.Server{
		Addr:    *addr,
		Handler: router,
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sc
		discord.Close()
		srv.Close()
	}()

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	log.Fatal(srv.ListenAndServe())
}
