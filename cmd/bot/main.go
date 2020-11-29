package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hiaooc/discgo/pkg/datastore"
	"github.com/hiaooc/discgo/pkg/handler"
)

var (
	token         = flag.String("token", os.Getenv("BOT_TOKEN"), "Bot token")
	dataStorePath = flag.String("datastore", "", "Path to JSON file")
	list          = flag.Bool("list", false, "Lists guilds and channels the bot is member of")
	guild         = flag.String("guild", os.Getenv("BOT_GUILD"), "Bot guild, currently only used for scheduled messages")
	channel       = flag.String("channel", os.Getenv("BOT_CHANNEL"), "Bot channel, currently only used for scheduled messages")
)

func logic() (*discordgo.Session, error) {
	if *token == "" {
		return nil, fmt.Errorf("Token needs to be configured via env variable BOT_TOKEN")
	}
	if *dataStorePath == "" {
		return nil, fmt.Errorf("-datastore must not be empty")
	}

	ds, err := datastore.Read(*dataStorePath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	dg, err := discordgo.New("Bot " + *token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}
	err = dg.Open()
	if err != nil {
		return nil, fmt.Errorf("create websocket connection: %w", err)
	}

	if *list {
		if err = listGuildsChannels(dg); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	dg.AddHandler(handler.ChangeTopic)
	dg.AddHandler(handler.PinMessage)
	dg.AddHandler(handler.UnpinMessage)

	replier := handler.NewReplier(ds)
	dg.AddHandler(replier.Handler)

	if *guild != "" && *channel != "" {
		go scheduledMessages(dg, *guild, *channel)
	}

	return dg, nil
}

func main() {
	flag.Parse()

	dg, err := logic()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-sc
	dg.Close()
}

func validateMembership(dg *discordgo.Session, guildID, channelID string) error {
	var guilds []string
	for _, g := range dg.State.Guilds {
		guild, err := dg.Guild(g.ID)
		if err != nil {
			return fmt.Errorf("get guild: %w", err)
		}

		if guildID == guild.ID {
			cc, err := dg.GuildChannels(guild.ID)
			if err != nil {
				return fmt.Errorf("get guild channels: %w", err)
			}

			var channels []string
			for _, c := range cc {
				if c.Type != discordgo.ChannelTypeGuildText {
					continue
				}
				if channelID == c.ID {
					return nil
				}
				channels = append(channels, fmt.Sprintf("%s (%s)", c.Name, c.ID))
			}

			return fmt.Errorf("could not find channel with ID %q, got: %s", channelID, strings.Join(channels, ", "))
		}
		guilds = append(guilds, fmt.Sprintf("%s (%s)", guild.Name, guild.ID))
	}

	return fmt.Errorf("could not find guild with ID %q, got: %s", guildID, strings.Join(guilds, ", "))
}

func scheduledMessages(dg *discordgo.Session, guildID, channelID string) {
	if err := validateMembership(dg, *guild, *channel); err != nil {
		log.Printf("deactivating scheduled messages: %v\n", err)
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case t := <-ticker.C:
			location, err := time.LoadLocation("Europe/Berlin")
			if err != nil {
				panic(err)
			}

			now := t.In(location)
			if now.Weekday() == time.Monday && now.Hour() == 10 && now.Minute() == 0 {
				_, err = dg.ChannelMessageSend(channelID, "**Worauf freut ihr euch diese Woche besonders?**")
				if err != nil {
					log.Println("send message: %v", err)
				}
			}
		}
	}
}

func listGuildsChannels(dg *discordgo.Session) error {
	for _, guild := range dg.State.Guilds {
		g, err := dg.Guild(guild.ID)
		if err != nil {
			return fmt.Errorf("get guild %q: %w", guild.ID, err)
		}

		log.Printf("%s (ID: %s)\n", g.Name, g.ID)
		cc, err := dg.GuildChannels(g.ID)
		if err != nil {
			return fmt.Errorf("get channels: %w", err)
		}

		for _, c := range cc {
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}
			log.Printf("- %s (ID: %s)\n", c.Name, c.ID)
		}
	}

	return nil
}
