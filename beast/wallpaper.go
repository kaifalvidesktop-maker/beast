package main

import "sync"

// ---------------------------------------------------
// HOME PAGE WALLPAPER / BACKGROUND CUSTOMIZATION
// ---------------------------------------------------

type WallpaperOption struct {
	ID       string
	Name     string
	CSSValue string
}

var wallpaperOptions = []WallpaperOption{
	{ID: "default", Name: "BEAST Dark", CSSValue: "linear-gradient(180deg, #0b0b0d 0%, #121214 100%)"},
	{ID: "midnight", Name: "Midnight Blue", CSSValue: "linear-gradient(180deg, #0a0e1a 0%, #131a2e 100%)"},
	{ID: "forest", Name: "Deep Forest", CSSValue: "linear-gradient(180deg, #0a120d 0%, #101c14 100%)"},
	{ID: "wine", Name: "Wine Red", CSSValue: "linear-gradient(180deg, #150a0d 0%, #1f1013 100%)"},
	{ID: "sunset", Name: "Dark Sunset", CSSValue: "linear-gradient(180deg, #1a0f0a 0%, #241610 100%)"},
	{ID: "solid", Name: "Pure Black", CSSValue: "#000000"},
}

type WallpaperManager struct {
	mu       sync.Mutex
	Selected string
}

var wallpaperManager = &WallpaperManager{
	Selected: "default",
}

// SetWallpaper changes the active wallpaper by ID
func (wm *WallpaperManager) SetWallpaper(id string) string {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, opt := range wallpaperOptions {
		if opt.ID == id {
			wm.Selected = id
			return wm.Selected
		}
	}
	return wm.Selected
}

// GetSelected returns the CSS value for the currently active wallpaper
func (wm *WallpaperManager) GetSelectedCSS() string {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for _, opt := range wallpaperOptions {
		if opt.ID == wm.Selected {
			return opt.CSSValue
		}
	}
	return wallpaperOptions[0].CSSValue
}

// GetAllOptions returns the full wallpaper list for the settings UI
func (wm *WallpaperManager) GetAllOptions() []WallpaperOption {
	return wallpaperOptions
}