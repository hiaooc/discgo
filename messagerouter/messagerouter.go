package messagerouter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Context struct {
	Content    string
	Fields     []string
	HasMention bool
	BotCommand bool
}

type HandlerFunc func(*discordgo.Session, *discordgo.Message, *Context)

type Route struct {
	Pattern string
	Handler HandlerFunc
}

func New() *Router {
	return &Router{}
}

type Router struct {
	Routes  []*Route
	Default *Route
}

func (r *Router) Route(pattern string, handler HandlerFunc) *Route {
	route := &Route{
		Pattern: pattern,
		Handler: handler,
	}
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	// fmt.Println("-------------")
	// Ignore messages by the bot
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	ctx := &Context{
		Content: strings.TrimSpace(mc.Content),
	}

	for _, mention := range mc.Mentions {
		if mention.ID == ds.State.User.ID {
			ctx.HasMention = true

			// If bot mention is first thing in message, it's probably a bot
			// command, remove the mentions from the message.
			reg := regexp.MustCompile(fmt.Sprintf("<@!%s>", ds.State.User.ID))
			if reg.FindStringIndex(ctx.Content)[0] == 0 {
				ctx.BotCommand = true
				ctx.Content = strings.TrimSpace(reg.ReplaceAllLiteralString(ctx.Content, ""))
			}
		}
	}

	fmt.Printf("ctx: %#v\n", ctx)

	route, fl := r.MatchRoute(ctx.Content)
	// fmt.Printf("route: %#v, fl: %#v\n", route, fl)
	if route != nil {
		ctx.Fields = fl
		route.Handler(ds, mc.Message, ctx)
		return
	}
}

func (r *Router) MatchRoute(msg string) (*Route, []string) {
	// Tokenize the message into words
	fields := strings.Fields(msg)

	// no point to continue if there's no fields
	if len(fields) == 0 {
		return nil, nil
	}

	// Search though the command list for a match
	var route *Route
	var rank int
	var fk int
	for fk, fv := range fields {
		for _, rv := range r.Routes {
			// If we find an exact match, return that immediately.
			if rv.Pattern == fv {
				return rv, fields[fk:]
			}

			// Some "Fuzzy" searching...
			if strings.HasPrefix(rv.Pattern, fv) {
				if len(fv) > rank {
					route = rv
					rank = len(fv)
				}
			}
		}
	}
	return route, fields[fk:]
}
