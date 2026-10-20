package main

import (
	"fmt"
	"minecraft-item-helper/internal/datafiles"
	"minecraft-item-helper/internal/planner"
	"sort"
	"strings"
)

const maxPageChars = 1800

type planSender interface {
	DeferPrivate() error
	SendPrivate(content string, page, pages int) error
	SetToken(string)
}
type planService struct {
	settings datafiles.Settings
	recipes  map[string]planner.Recipe
	tags     planner.Tags
	newPager func(owner string, pages []string) (string, int, error)
}

func (s planService) handle(item string, amount int64, owner string, sender planSender) error {
	if err := sender.DeferPrivate(); err != nil {
		return err
	}
	if amount <= 0 || strings.TrimSpace(item) == "" || strings.HasPrefix(item, "#") {
		return sender.SendPrivate("Enter a concrete item ID and a quantity greater than zero.", 0, 0)
	}
	response, err := planner.PlanBuild(planner.BuildRequest{Items: map[string]uint{item: uint(amount)}, Recipes: s.recipes, Tags: s.tags, StopItems: s.settings.StopItemSet(), Preferences: s.settings.Preferences})
	if err != nil {
		message := "I couldn't plan that request: " + err.Error()
		if len(message) > maxPageChars {
			message = message[:maxPageChars-3] + "..."
		}
		return sender.SendPrivate(message, 0, 0)
	}
	pages := materialPages(response.FlatMap)
	if len(pages) <= 1 {
		return sender.SendPrivate(pages[0], 0, 0)
	}
	if s.newPager != nil {
		token, page, e := s.newPager(owner, pages)
		if e != nil {
			return sender.SendPrivate("The plan is too long to display right now. Please try again.", 0, 0)
		}
		sender.SetToken(token)
		return sender.SendPrivate(pages[page], page, len(pages))
	}
	for i, page := range pages {
		if err := sender.SendPrivate(page, i, len(pages)); err != nil {
			return err
		}
	}
	return nil
}
func materialPages(materials map[string]uint) []string {
	ids := make([]string, 0, len(materials))
	for id := range materials {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var pages []string
	current := "Materials to gather:\n"
	for _, id := range ids {
		line := fmt.Sprintf("%s: %d\n", id, materials[id])
		if len(current)+len(line) > maxPageChars && current != "Materials to gather:\n" {
			pages = append(pages, strings.TrimRight(current, "\n"))
			current = ""
		}
		current += line
	}
	if current == "Materials to gather:\n" {
		return []string{"No additional materials needed."}
	}
	return append(pages, strings.TrimRight(current, "\n"))
}
