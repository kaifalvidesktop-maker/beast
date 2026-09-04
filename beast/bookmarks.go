package main

import (
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// BOOKMARK SYSTEM (RAM ONLY)
// ---------------------------------------------------

type Bookmark struct {
	ID        int
	Title     string
	URL       string
	Folder    string
	CreatedAt time.Time
}

type BookmarkManager struct {
	mu        sync.Mutex
	Bookmarks []*Bookmark
	NextID    int
}

var bookmarkManager = &BookmarkManager{
	Bookmarks: []*Bookmark{},
	NextID:    1,
}

// Add a new bookmark
func (bm *BookmarkManager) Add(title string, url string, folder string) *Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if folder == "" {
		folder = "Uncategorized"
	}

	b := &Bookmark{
		ID:        bm.NextID,
		Title:     title,
		URL:       url,
		Folder:    folder,
		CreatedAt: time.Now(),
	}

	bm.Bookmarks = append(bm.Bookmarks, b)
	bm.NextID++

	return b
}

// Remove a bookmark by ID
func (bm *BookmarkManager) Remove(id int) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	newList := []*Bookmark{}
	removed := false

	for _, b := range bm.Bookmarks {
		if b.ID == id {
			removed = true
			continue
		}
		newList = append(newList, b)
	}

	bm.Bookmarks = newList
	return removed
}

// Remove a bookmark by URL (used for quick toggle)
func (bm *BookmarkManager) RemoveByURL(url string) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	newList := []*Bookmark{}
	removed := false

	for _, b := range bm.Bookmarks {
		if b.URL == url {
			removed = true
			continue
		}
		newList = append(newList, b)
	}

	bm.Bookmarks = newList
	return removed
}

// Check if a URL is already bookmarked
func (bm *BookmarkManager) IsBookmarked(url string) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, b := range bm.Bookmarks {
		if b.URL == url {
			return true
		}
	}
	return false
}

// Toggle bookmark: add if not present, remove if present
func (bm *BookmarkManager) Toggle(title string, url string) bool {
	if bm.IsBookmarked(url) {
		bm.RemoveByURL(url)
		return false // now NOT bookmarked
	}
	bm.Add(title, url, "Uncategorized")
	return true // now bookmarked
}

// Get all bookmarks
func (bm *BookmarkManager) GetAll() []*Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.Bookmarks
}

// Search bookmarks by title or URL keyword
func (bm *BookmarkManager) Search(keyword string) []*Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	keyword = strings.ToLower(keyword)
	results := []*Bookmark{}

	for _, b := range bm.Bookmarks {
		if strings.Contains(strings.ToLower(b.Title), keyword) ||
			strings.Contains(strings.ToLower(b.URL), keyword) {
			results = append(results, b)
		}
	}
	return results
}

// Get bookmarks grouped by folder (for UI sidebar)
func (bm *BookmarkManager) GetFolders() map[string][]*Bookmark {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	grouped := make(map[string][]*Bookmark)
	for _, b := range bm.Bookmarks {
		grouped[b.Folder] = append(grouped[b.Folder], b)
	}
	return grouped

}