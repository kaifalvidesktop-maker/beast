package main

import (
	"sync"
	"time"
)

// ---------------------------------------------------
// TAB SYSTEM (RAM ONLY - MULTIPLE TAB TRACKING)
// ---------------------------------------------------

type Tab struct {
	ID        int
	Title     string
	URL       string
	CreatedAt time.Time
	IsActive  bool
}

type TabManager struct {
	mu       sync.Mutex
	Tabs     []*Tab
	NextID   int
	ActiveID int
}

var tabManager = &TabManager{
	Tabs:     []*Tab{},
	NextID:   1,
	ActiveID: 0,
}

// Create a new tab
func (tm *TabManager) NewTab(url string) *Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Deactivate all existing tabs
	for _, t := range tm.Tabs {
		t.IsActive = false
	}

	tab := &Tab{
		ID:        tm.NextID,
		Title:     "New Tab",
		URL:       url,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	tm.Tabs = append(tm.Tabs, tab)
	tm.ActiveID = tab.ID
	tm.NextID++

	return tab
}

// Close a tab by ID
func (tm *TabManager) CloseTab(id int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	newTabs := []*Tab{}
	wasActive := false

	for _, t := range tm.Tabs {
		if t.ID == id {
			if t.IsActive {
				wasActive = true
			}
			continue
		}
		newTabs = append(newTabs, t)
	}

	tm.Tabs = newTabs

	// If the closed tab was active, activate the last remaining tab
	if wasActive && len(tm.Tabs) > 0 {
		lastTab := tm.Tabs[len(tm.Tabs)-1]
		lastTab.IsActive = true
		tm.ActiveID = lastTab.ID
	}
}

// Switch to a specific tab
func (tm *TabManager) SwitchTab(id int) *Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var found *Tab
	for _, t := range tm.Tabs {
		if t.ID == id {
			t.IsActive = true
			found = t
			tm.ActiveID = id
		} else {
			t.IsActive = false
		}
	}
	return found
}

// Update a tab's title and URL after navigation
func (tm *TabManager) UpdateTab(id int, title string, url string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.Tabs {
		if t.ID == id {
			t.Title = title
			t.URL = url
			break
		}
	}
}

// Get the currently active tab
func (tm *TabManager) GetActiveTab() *Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.Tabs {
		if t.ID == tm.ActiveID {
			return t
		}
	}
	return nil
}

// Get all tabs (for rendering tab bar UI)
func (tm *TabManager) GetAllTabs() []*Tab {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.Tabs
}

// Close all tabs (used on full reset / all-delete)
func (tm *TabManager) CloseAllTabs() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Tabs = []*Tab{}
	tm.ActiveID = 0
}