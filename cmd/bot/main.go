package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/hiaooc/discgo/pkg/datastore"
	"github.com/hiaooc/discgo/pkg/handler"
)

var (
	dataStorePath = flag.String("datastore", "", "Path to JSON file")
)

func main() {
	flag.Parse()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Token needs to be configured via env variable BOT_TOKEN")
	}
	if *dataStorePath == "" {
		log.Fatal("-datastore must not be empty")
	}

	ds, err := datastore.Read(*dataStorePath)
	if err != nil {
		log.Fatal(err)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(handler.ChangeTopic)
	dg.AddHandler(handler.PinMessage)
	dg.AddHandler(handler.UnpinMessage)

	replier := handler.NewReplier(ds)
	dg.AddHandler(replier.Handler)

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	go scheduledMessages()

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()
}

func scheduledMessages() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			location, err := time.LoadLocation("Europe/Berlin")
			if err != nil {
				panic(err)
			}

			now := time.Now().In(location)
			if now.Weekday() == time.Monday && now.Hour() == 10 && now.Minute() == 0 {
				// Worauf freut ihr euch diese Woche besonders?
			}
		}
	}
}
