package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"minecraft-item-helper/internal/datafiles"
)

type runtime struct {
	service planService
	store   *pageStore
}

func intPointer(v int) *int { return &v }

func isPlanCommandPath(path string) bool { return path == "/plan" }

func env(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
func loadRuntime() (runtime, error) {
	settings, e := datafiles.LoadSettings(os.Getenv("MINECRAFT_ITEM_HELPER_CONFIG"))
	if e != nil {
		return runtime{}, e
	}
	recipes, e := datafiles.LoadRecipes(env("MINECRAFT_ITEM_HELPER_RECIPES", "assets/recipes"), settings.RecipeChoices)
	if e != nil {
		return runtime{}, e
	}
	tags, e := datafiles.LoadTags(env("MINECRAFT_ITEM_HELPER_TAGS", "assets/item-tag-groups"))
	if e != nil {
		return runtime{}, e
	}
	return runtime{service: planService{settings: settings, recipes: recipes, tags: tags}, store: newPageStore()}, nil
}

type discordSender struct {
	event *events.ApplicationCommandInteractionCreate
	token string
}

func (s *discordSender) SetToken(token string) { s.token = token }
func (s *discordSender) DeferPrivate() error   { return s.event.DeferCreateMessage(true) }
func pageContent(content string, page, pages int) string {
	if pages > 1 {
		return fmt.Sprintf("%s\nPage %d of %d", content, page+1, pages)
	}
	return content
}
func (s *discordSender) SendPrivate(content string, page, pages int) error {
	content = pageContent(content, page, pages)
	if page == 0 {
		update := discord.NewMessageUpdate().WithContent(content)
		if s.token != "" && pages > 1 {
			update = update.WithComponents(pageButtons(s.token, page, pages))
		}
		_, e := s.event.Client().Rest.UpdateInteractionResponse(s.event.ApplicationID(), s.event.Token(), update)
		return e
	}
	msg := discord.NewMessageCreate().WithContent(content).WithEphemeral(true).WithComponents(pageButtons(s.token, page, pages))
	_, e := s.event.Client().Rest.CreateFollowupMessage(s.event.ApplicationID(), s.event.Token(), msg)
	return e
}
func pageButtons(token string, page, pages int) discord.ActionRowComponent {
	return discord.NewActionRow(discord.NewSecondaryButton("Previous", "plan:prev:"+token).WithDisabled(page == 0), discord.NewSecondaryButton("Next", "plan:next:"+token).WithDisabled(page == pages-1))
}
func main() {
	if os.Getenv("DISCORD_BOT_TOKEN") == "" {
		log.Fatal("DISCORD_BOT_TOKEN is required")
	}
	rt, e := loadRuntime()
	if e != nil {
		log.Fatalf("load planning data: %v", e)
	}
	rt.service.newPager = func(owner string, pages []string) (string, int, error) {
		token, e := rt.store.create(owner, pages)
		return token, 0, e
	}
	listener := func(e *events.ApplicationCommandInteractionCreate) {
		data := e.SlashCommandInteractionData()
		if !isPlanCommandPath(data.CommandPath()) {
			return
		}
		log.Printf("received /plan interaction")
		sender := &discordSender{event: e}
		if err := rt.service.handle(data.String("item"), int64(data.Int("quantity")), e.User().ID.String(), sender); err != nil {
			log.Printf("failed to respond to planning interaction: %v", err)
		}
	}
	autocomplete := func(e *events.AutocompleteInteractionCreate) {
		data := e.Data
		focused := data.Focused()
		if focused.Name != "item" {
			_ = e.AutocompleteResult(nil)
			return
		}
		matches := suggestions(rt.service.recipes, rt.service.tags, focused.String())
		choices := make([]discord.AutocompleteChoice, 0, len(matches))
		for _, m := range matches {
			choices = append(choices, discord.AutocompleteChoiceString{Name: m.Name, Value: m.Value})
		}
		_ = e.AutocompleteResult(choices)
	}
	components := func(e *events.ComponentInteractionCreate) {
		parts := strings.Split(e.ButtonInteractionData().CustomID(), ":")
		if len(parts) != 3 || parts[0] != "plan" {
			return
		}
		if parts[1] != "next" && parts[1] != "prev" {
			return
		}
		delta := -1
		if parts[1] == "next" {
			delta = 1
		}
		page, index, total, err := rt.store.move(parts[2], e.User().ID.String(), delta)
		if err != nil {
			_ = e.CreateMessage(discord.NewMessageCreate().WithContent("This plan expired or this control is not yours.").WithEphemeral(true))
			return
		}
		_ = e.UpdateMessage(discord.NewMessageUpdate().WithContent(pageContent(page, index, total)).WithComponents(pageButtons(parts[2], index, total)))
	}
	client, e := disgo.New(os.Getenv("DISCORD_BOT_TOKEN"), bot.WithDefaultGateway(), bot.WithEventListenerFunc(listener), bot.WithEventListenerFunc(autocomplete), bot.WithEventListenerFunc(components))
	if e != nil {
		log.Fatalf("create Discord client: %v", e)
	}
	defer client.Close(context.Background())
	maxItemLength := 256
	command := discord.SlashCommandCreate{Name: "plan", Description: "Plan materials for a Minecraft item", Options: []discord.ApplicationCommandOption{discord.ApplicationCommandOptionString{Name: "item", Description: "Minecraft item", Required: true, Autocomplete: true, MaxLength: &maxItemLength}, discord.ApplicationCommandOptionInt{Name: "quantity", Description: "Number of items", Required: true, MinValue: intPointer(1)}}}
	if _, e = client.Rest.SetGlobalCommands(client.ApplicationID, []discord.ApplicationCommandCreate{command}); e != nil {
		log.Fatalf("register /plan: %v", e)
	}
	log.Printf("registered global Discord command /plan")
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rt.store.mu.Lock()
			rt.store.prune()
			rt.store.mu.Unlock()
		}
	}()
	if e = client.OpenGateway(context.Background()); e != nil {
		log.Fatalf("open Discord gateway: %v", e)
	}
	log.Printf("Discord Gateway connected")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}
