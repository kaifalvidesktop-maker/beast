package main

import (
	"strings"
	"sync"
)

// ---------------------------------------------------
// SETTINGS SYSTEM (RAM ONLY - RESETS ON RESTART)
// ---------------------------------------------------

type BrowserSettings struct {
	mu                  sync.Mutex
	Theme               string // "dark" or "light"
	DefaultSearchEngine string // "google", "bing", "duckduckgo"
	Homepage            string
	JSEnabled           bool
	ImagesEnabled       bool
	DownloadPath        string
}

var settings = &BrowserSettings{
	Theme:               "dark",
	DefaultSearchEngine: "google",
	Homepage:            "beast://home",
	JSEnabled:           true,
	ImagesEnabled:       true,
	DownloadPath:        "Downloads",
}

// Get all settings as a map (for sending to UI)
func (s *BrowserSettings) GetAll() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"theme":         s.Theme,
		"searchEngine":  s.DefaultSearchEngine,
		"homepage":      s.Homepage,
		"jsEnabled":     s.JSEnabled,
		"imagesEnabled": s.ImagesEnabled,
		"downloadPath":  s.DownloadPath,
	}
}

// Change theme (dark/light)
func (s *BrowserSettings) SetTheme(theme string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	theme = strings.ToLower(strings.TrimSpace(theme))
	if theme != "dark" && theme != "light" {
		theme = "dark"
	}
	s.Theme = theme
	return s.Theme
}

// Change default search engine
func (s *BrowserSettings) SetSearchEngine(engine string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	engine = strings.ToLower(strings.TrimSpace(engine))
	valid := map[string]bool{"google": true, "bing": true, "duckduckgo": true}

	if !valid[engine] {
		engine = "google"
	}
	s.DefaultSearchEngine = engine
	return s.DefaultSearchEngine
}

// Change homepage URL
func (s *BrowserSettings) SetHomepage(url string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	url = strings.TrimSpace(url)
	if url == "" {
		url = "beast://home"
	}
	s.Homepage = url
	return s.Homepage
}

// Toggle JavaScript on/off
func (s *BrowserSettings) ToggleJS() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.JSEnabled = !s.JSEnabled
	return s.JSEnabled
}

// Toggle Image loading on/off
func (s *BrowserSettings) ToggleImages() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ImagesEnabled = !s.ImagesEnabled
	return s.ImagesEnabled
}

// Change download folder path
func (s *BrowserSettings) SetDownloadPath(path string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	path = strings.TrimSpace(path)
	if path == "" {
		path = "Downloads"
	}
	s.DownloadPath = path
	return s.DownloadPath
}

// Reset everything back to default values
func (s *BrowserSettings) ResetToDefault() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Theme = "dark"
	s.DefaultSearchEngine = "google"
	s.Homepage = "beast://home"
	s.JSEnabled = true
	s.ImagesEnabled = true
	s.DownloadPath = "Downloads"

	return map[string]any{
		"theme":         s.Theme,
		"searchEngine":  s.DefaultSearchEngine,
		"homepage":      s.Homepage,
		"jsEnabled":     s.JSEnabled,
		"imagesEnabled": s.ImagesEnabled,
		"downloadPath":  s.DownloadPath,
	}
}

// Build a search URL based on the currently selected search engine
func buildSearchURL(query string) string {
	query = strings.ReplaceAll(strings.TrimSpace(query), " ", "+")

	switch settings.DefaultSearchEngine {
	case "bing":
		return "https://www.bing.com/search?q=" + query
	case "duckduckgo":
		return "https://duckduckgo.com/?q=" + query
	default:
		return "https://www.google.com/search?q=" + query
	}
}
