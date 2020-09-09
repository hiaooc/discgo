package main

import (
	"flag"
	"fmt"
	"log"
	ghttp "net/http"
	"os"
	"os/signal"
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

	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		handler := http.NewHandler(ds)
		r.Get("/config", handler.ReadConfig)
		r.Post("/config", handler.WriteConfig)
	})

	r.Route("/slackbot", func(r chi.Router) {
		r.Get("/view", bot.ViewHandler)
		r.Get("/edit", bot.EditHandler)
	})

	srv := ghttp.Server{
		Addr:    *addr,
		Handler: r,
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
