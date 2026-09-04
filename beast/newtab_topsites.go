package main

import (
	"sort"
	"strings"
	"sync"
)

// ---------------------------------------------------
// SMART TOP SITES (based on visit frequency from history,
// falls back to defaults when history is empty)
// ---------------------------------------------------

type TopSite struct {
	Domain string
	Visits int
	URL    string
}

type TopSitesEngine struct {
	mu sync.Mutex
}

var topSitesEngine = &TopSitesEngine{}

var defaultTopSites = []TopSite{
	{Domain: "youtube.com", URL: "https://youtube.com"},
	{Domain: "facebook.com", URL: "https://facebook.com"},
	{Domain: "github.com", URL: "https://github.com"},
	{Domain: "chat.openai.com", URL: "https://chat.openai.com"},
}

// GetTopSites returns most-visited domains from history, or defaults
func (te *TopSitesEngine) GetTopSites(limit int) []TopSite {
	te.mu.Lock()
	defer te.mu.Unlock()

	entries := history.GetAll()
	if len(entries) == 0 {
		return capSites(defaultTopSites, limit)
	}

	counts := make(map[string]int)
	firstURL := make(map[string]string)

	for _, e := range entries {
		domain := extractDomain(e.URL)
		if domain == "" {
			continue
		}
		counts[domain]++
		if _, exists := firstURL[domain]; !exists {
			firstURL[domain] = e.URL
		}
	}

	sites := make([]TopSite, 0, len(counts))
	for domain, count := range counts {
		sites = append(sites, TopSite{
			Domain: domain,
			Visits: count,
			URL:    firstURL[domain],
		})
	}

	sort.SliceStable(sites, func(i, j int) bool {
		return sites[i].Visits > sites[j].Visits
	})

	if len(sites) == 0 {
		return capSites(defaultTopSites, limit)
	}

	return capSites(sites, limit)
}

func capSites(sites []TopSite, limit int) []TopSite {
	if len(sites) > limit {
		return sites[:limit]
	}
	return sites
}

// PinnedTopSite lets a user manually pin a shortcut on the new tab page
type PinnedShortcut struct {
	Title string
	URL   string
}

type ShortcutManager struct {
	mu        sync.Mutex
	Shortcuts []PinnedShortcut
}

var shortcutManagerNewTab = &ShortcutManager{
	Shortcuts: []PinnedShortcut{},
}

// AddShortcut pins a custom shortcut to the new tab page
func (sm *ShortcutManager) AddShortcut(title string, url string) []PinnedShortcut {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	title = strings.TrimSpace(title)
	if title == "" {
		title = extractDomain(url)
	}

	sm.Shortcuts = append(sm.Shortcuts, PinnedShortcut{Title: title, URL: url})

	if len(sm.Shortcuts) > 10 {
		sm.Shortcuts = sm.Shortcuts[len(sm.Shortcuts)-10:]
	}
	return sm.Shortcuts
}

// RemoveShortcut removes a pinned shortcut by index
func (sm *ShortcutManager) RemoveShortcut(index int) []PinnedShortcut {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if index < 0 || index >= len(sm.Shortcuts) {
		return sm.Shortcuts
	}

	newList := []PinnedShortcut{}
	for i, s := range sm.Shortcuts {
		if i != index {
			newList = append(newList, s)
		}
	}
	sm.Shortcuts = newList
	return sm.Shortcuts
}

// GetShortcuts returns all pinned shortcuts
func (sm *ShortcutManager) GetShortcuts() []PinnedShortcut {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.Shortcuts
}