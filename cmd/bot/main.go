package main

import (
	"flag"
	"fmt"
	"log"
	ghttp "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
	"github.com/hiaooc/discgo/pkg/datastore"
	"github.com/hiaooc/discgo/pkg/handler"
	"github.com/hiaooc/discgo/pkg/http"
	"github.com/hiaooc/discgo/pkg/slackbot"
)

var (
	dataStorePath = flag.String("datastore", "", "Path to JSON file")
	addr          = flag.String("listen", ":8080", "Listen address")
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
	discord.AddHandler(handler.ChangeTopic)

	discord.AddHandler(handler.PinMessage)
	discord.AddHandler(handler.UnpinMessage)

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
