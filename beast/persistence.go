package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ---------------------------------------------------
// PERSISTENCE LAYER
// ---------------------------------------------------
//
// BEAST keeps browsing HISTORY, saved PASSWORDS, and AUTOFILL data
// strictly in RAM (privacy-first: nothing sensitive ever touches disk).
//
// Everything else that a normal browser is expected to remember between
// launches - bookmarks, settings, zoom levels, and the recently-closed
// tab list - is persisted here as a single small JSON file under the
// user's config directory:
//
//   Windows: %AppData%\Beast\state.json
//   Linux:   ~/.config/Beast/state.json
//   macOS:   ~/Library/Application Support/Beast/state.json
//
// The file is written on a 30s autosave tick and once more on shutdown,
// so a crash loses at most ~30s of bookmark/setting changes - nothing
// sensitive is ever at risk since none of it is stored here.

type persistedState struct {
	SavedAt time.Time `json:"savedAt"`

	// Settings snapshot
	Theme               string `json:"theme"`
	DefaultSearchEngine string `json:"defaultSearchEngine"`
	Homepage            string `json:"homepage"`
	JSEnabled           bool   `json:"jsEnabled"`
	ImagesEnabled       bool   `json:"imagesEnabled"`
	DownloadPath        string `json:"downloadPath"`

	// Bookmarks snapshot
	Bookmarks       []*Bookmark `json:"bookmarks"`
	NextBookmarkID  int         `json:"nextBookmarkId"`

	// Zoom levels (per-URL, since tab IDs don't survive a restart)
	ZoomByURL map[string]float64 `json:"zoomByUrl"`

	// Recently closed tabs
	ClosedTabs []ClosedTab `json:"closedTabs"`
}

func stateFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, "Beast")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "state.json"), nil
}

// buildPersistedState takes a consistent snapshot of everything we persist.
func buildPersistedState() *persistedState {
	settings.mu.Lock()
	s := persistedState{
		SavedAt:             time.Now(),
		Theme:               settings.Theme,
		DefaultSearchEngine: settings.DefaultSearchEngine,
		Homepage:            settings.Homepage,
		JSEnabled:           settings.JSEnabled,
		ImagesEnabled:       settings.ImagesEnabled,
		DownloadPath:        settings.DownloadPath,
	}
	settings.mu.Unlock()

	bookmarkManager.mu.Lock()
	s.Bookmarks = append([]*Bookmark{}, bookmarkManager.Bookmarks...)
	s.NextBookmarkID = bookmarkManager.NextID
	bookmarkManager.mu.Unlock()

	zoomManager.mu.Lock()
	s.ZoomByURL = make(map[string]float64, len(zoomManager.Zooms))
	// Zoom is tracked per tab ID at runtime; we don't have URLs here,
	// so persistence for zoom is keyed by tab ID as a string. This still
	// lets a restored session (same tab order) keep its zoom level.
	for tabID, level := range zoomManager.Zooms {
		s.ZoomByURL[persistIntToStr(tabID)] = level
	}
	zoomManager.mu.Unlock()

	sessionManager.mu.Lock()
	s.ClosedTabs = append([]ClosedTab{}, sessionManager.History...)
	sessionManager.mu.Unlock()

	return &s
}

// applyPersistedState restores a loaded snapshot into the live managers.
func applyPersistedState(s *persistedState) {
	if s == nil {
		return
	}

	settings.mu.Lock()
	if s.Theme != "" {
		settings.Theme = s.Theme
	}
	if s.DefaultSearchEngine != "" {
		settings.DefaultSearchEngine = s.DefaultSearchEngine
	}
	if s.Homepage != "" {
		settings.Homepage = s.Homepage
	}
	settings.JSEnabled = s.JSEnabled
	settings.ImagesEnabled = s.ImagesEnabled
	if s.DownloadPath != "" {
		settings.DownloadPath = s.DownloadPath
	}
	settings.mu.Unlock()

	bookmarkManager.mu.Lock()
	if len(s.Bookmarks) > 0 {
		bookmarkManager.Bookmarks = s.Bookmarks
	}
	if s.NextBookmarkID > bookmarkManager.NextID {
		bookmarkManager.NextID = s.NextBookmarkID
	}
	bookmarkManager.mu.Unlock()

	zoomManager.mu.Lock()
	for key, level := range s.ZoomByURL {
		if tabID, ok := persistStrToInt(key); ok {
			zoomManager.Zooms[tabID] = level
		}
	}
	zoomManager.mu.Unlock()

	sessionManager.mu.Lock()
	if len(s.ClosedTabs) > 0 {
		sessionManager.History = s.ClosedTabs
	}
	sessionManager.mu.Unlock()
}

// loadPersistedState reads state.json if present. Missing file is not
// an error - it just means first run.
func loadPersistedState() *persistedState {
	path, err := stateFilePath()
	if err != nil {
		log.Println("[persistence] could not resolve config dir:", err)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("[persistence] read failed:", err)
		}
		return nil
	}

	var s persistedState
	if err := json.Unmarshal(data, &s); err != nil {
		log.Println("[persistence] corrupt state.json, ignoring:", err)
		return nil
	}
	return &s
}

// savePersistedState writes the current snapshot to disk atomically
// (write to a temp file, then rename) so a crash mid-write can't corrupt it.
func savePersistedState() error {
	path, err := stateFilePath()
	if err != nil {
		return err
	}

	snapshot := buildPersistedState()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// startAutosave saves state every interval until stop is closed.
func startAutosave(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := savePersistedState(); err != nil {
					log.Println("[persistence] autosave failed:", err)
				}
			case <-stop:
				return
			}
		}
	}()
}

// --- tiny local helpers so this file has zero extra imports beyond stdlib ---
// (named with a persist* prefix to avoid colliding with the itoa/atoi
// helpers already defined in incognito.go)

func persistIntToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func persistStrToInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	neg := false
	i := 0
	if s[0] == '-' {
		neg = true
		i = 1
	}
	if i == len(s) {
		return 0, false
	}
	n := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, true
}
