package main

import (
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// DOWNLOAD TRACKER (RAM ONLY - LIST CLEARS ON RESTART)
// ---------------------------------------------------

type DownloadItem struct {
	ID         int
	FileName   string
	URL        string
	Progress   int // 0 to 100
	Status     string // "downloading", "completed", "failed", "cancelled"
	StartedAt  time.Time
	FinishedAt time.Time
}

type DownloadManager struct {
	mu     sync.Mutex
	Items  []*DownloadItem
	NextID int
}

var downloadManager = &DownloadManager{
	Items:  []*DownloadItem{},
	NextID: 1,
}

// Start tracking a new download
func (dm *DownloadManager) Start(url string) *DownloadItem {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	fileName := filepath.Base(url)
	if fileName == "" || fileName == "." || strings.Contains(fileName, "?") {
		parts := strings.Split(fileName, "?")
		fileName = parts[0]
	}
	if fileName == "" {
		fileName = "unknown_file"
	}

	item := &DownloadItem{
		ID:        dm.NextID,
		FileName:  fileName,
		URL:       url,
		Progress:  0,
		Status:    "downloading",
		StartedAt: time.Now(),
	}

	dm.Items = append(dm.Items, item)
	dm.NextID++

	return item
}

// Update progress percentage of a download
func (dm *DownloadManager) UpdateProgress(id int, progress int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}

	for _, item := range dm.Items {
		if item.ID == id {
			item.Progress = progress
			break
		}
	}
}

// Mark a download as completed
func (dm *DownloadManager) Complete(id int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, item := range dm.Items {
		if item.ID == id {
			item.Status = "completed"
			item.Progress = 100
			item.FinishedAt = time.Now()
			break
		}
	}
}

// Mark a download as failed
func (dm *DownloadManager) Fail(id int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, item := range dm.Items {
		if item.ID == id {
			item.Status = "failed"
			item.FinishedAt = time.Now()
			break
		}
	}
}

// Cancel an in-progress download
func (dm *DownloadManager) Cancel(id int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, item := range dm.Items {
		if item.ID == id {
			item.Status = "cancelled"
			item.FinishedAt = time.Now()
			break
		}
	}
}

// Remove all completed/failed/cancelled downloads from the list
func (dm *DownloadManager) ClearFinished() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	active := []*DownloadItem{}
	for _, item := range dm.Items {
		if item.Status == "downloading" {
			active = append(active, item)
		}
	}
	dm.Items = active
}

// Get all downloads (active + finished)
func (dm *DownloadManager) GetAll() []*DownloadItem {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.Items
}