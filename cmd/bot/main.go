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

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	dg.Close()
}
