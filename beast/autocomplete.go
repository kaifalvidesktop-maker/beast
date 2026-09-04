package main

import (
	"sort"
	"strings"
	"sync"
)

// ---------------------------------------------------
// ADDRESS BAR AUTOCOMPLETE / SUGGESTION ENGINE
// Combines History + Bookmarks + common site guesses
// ---------------------------------------------------

type Suggestion struct {
	Text   string
	URL    string
	Source string // "history", "bookmark", "search"
}

var commonSites = []string{
	"youtube.com", "facebook.com", "github.com", "chat.openai.com",
	"google.com", "wikipedia.org", "reddit.com", "twitter.com",
	"instagram.com", "amazon.com", "netflix.com", "stackoverflow.com",
}

type SuggestionEngine struct {
	mu sync.Mutex
}

var suggestEngine = &SuggestionEngine{}

// GetSuggestions returns up to `limit` ranked suggestions for a partial input
func (se *SuggestionEngine) GetSuggestions(partial string, limit int) []Suggestion {
	se.mu.Lock()
	defer se.mu.Unlock()

	partial = strings.ToLower(strings.TrimSpace(partial))
	if partial == "" {
		return []Suggestion{}
	}

	results := []Suggestion{}
	seen := make(map[string]bool)

	// 1. Bookmarks (highest priority - user explicitly saved these)
	for _, b := range bookmarkManager.GetAll() {
		lowerTitle := strings.ToLower(b.Title)
		lowerURL := strings.ToLower(b.URL)
		if strings.Contains(lowerTitle, partial) || strings.Contains(lowerURL, partial) {
			if !seen[b.URL] {
				results = append(results, Suggestion{Text: b.Title, URL: b.URL, Source: "bookmark"})
				seen[b.URL] = true
			}
		}
	}

	// 2. History (recent visits matching)
	recent := history.GetRecent(300)
	for i := len(recent) - 1; i >= 0; i-- {
		entry := recent[i]
		lowerURL := strings.ToLower(entry.URL)
		if strings.Contains(lowerURL, partial) && !seen[entry.URL] {
			results = append(results, Suggestion{Text: entry.URL, URL: entry.URL, Source: "history"})
			seen[entry.URL] = true
		}
		if len(results) >= limit*2 {
			break
		}
	}

	// 3. Common site guesses (e.g. typing "you" suggests youtube.com)
	for _, site := range commonSites {
		if strings.HasPrefix(site, partial) && !seen[site] {
			full := "https://" + site
			results = append(results, Suggestion{Text: site, URL: full, Source: "search"})
			seen[site] = true
		}
	}

	// 4. Fallback: plain Google search suggestion always included
	results = append(results, Suggestion{
		Text:   "Search Google for \"" + partial + "\"",
		URL:    partial,
		Source: "search",
	})

	sortSuggestionsByRelevance(results, partial)

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

// sortSuggestionsByRelevance ranks bookmarks first, then shorter matches
func sortSuggestionsByRelevance(list []Suggestion, query string) {
	query=strings.ToLower(query)
	sort.SliceStable(list, func(i, j int) bool {
		weightI := sourceWeight(list[i].Source)
		weightJ := sourceWeight(list[j].Source)
		if weightI != weightJ {
			return weightI < weightJ
		}
		return len(list[i].Text) < len(list[j].Text)
	})
}

func sourceWeight(source string) int {
	switch source {
	case "bookmark":
		return 0
	case "history":
		return 1
	default:
		return 2
	}
}