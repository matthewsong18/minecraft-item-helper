package main

import (
	"minecraft-item-helper/internal/planner"
	"sort"
	"strings"
	"unicode"
)

type itemSuggestion struct{ Name, Value string }

func suggestions(recipes map[string]planner.Recipe, tags planner.Tags, query string) []itemSuggestion {
	items := map[string]bool{}
	for output, r := range recipes {
		items[output] = true
		for _, in := range r.Inputs {
			for _, id := range in.Options {
				if !strings.HasPrefix(id, "#") {
					items[id] = true
				}
			}
		}
	}
	var expand func(string, map[string]bool)
	expand = func(tag string, seen map[string]bool) {
		if seen[tag] {
			return
		}
		seen[tag] = true
		for _, v := range tags[tag] {
			if strings.HasPrefix(v, "#") {
				expand(v, seen)
			} else {
				items[v] = true
			}
		}
	}
	for tag := range tags {
		expand(tag, map[string]bool{})
	}
	q := strings.ToLower(strings.TrimSpace(query))
	ids := make([]string, 0, len(items))
	for id := range items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var prefix, contains []itemSuggestion
	for _, id := range ids {
		name := prettyID(id)
		label := name + " (" + id + ")"
		if len(label) > 100 {
			label = label[:97] + "..."
		}
		if q == "" || strings.HasPrefix(strings.ToLower(name), q) || strings.HasPrefix(strings.ToLower(id), q) {
			prefix = append(prefix, itemSuggestion{Name: label, Value: id})
		} else if strings.Contains(strings.ToLower(name+" "+id), q) {
			contains = append(contains, itemSuggestion{Name: label, Value: id})
		}
	}
	result := append(prefix, contains...)
	if len(result) > 25 {
		result = result[:25]
	}
	return result
}
func prettyID(id string) string {
	id = strings.TrimPrefix(id, "minecraft:")
	var b strings.Builder
	upper := true
	for _, r := range id {
		if r == '_' || r == '-' {
			b.WriteRune(' ')
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
