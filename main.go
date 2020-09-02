package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bl1nk/discgo/datastore"
	"github.com/bl1nk/discgo/slackbot"
	"github.com/bwmarrin/discordgo"
)

var dataStorePath = flag.String("datastore", "", "Path to JSON file")

func main() {
	flag.Parse()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Token needs to be configured via env variable BOT_TOKEN")
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

	err = discord.Open()

	if err != nil {
		log.Fatal(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	_ = discord.Close()

	fmt.Println("Test")
}
