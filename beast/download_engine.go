package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ---------------------------------------------------
// REAL DOWNLOAD ENGINE (actually saves files to disk)
// ---------------------------------------------------

type ActiveDownload struct {
	mu       sync.Mutex
	Item     *DownloadItem
	Cancelled bool
}

var activeDownloads = struct {
	mu    sync.Mutex
	Items map[int]*ActiveDownload
}{Items: make(map[int]*ActiveDownload)}

// StartDownload kicks off a real HTTP download in a background goroutine
func StartDownload(rawURL string, saveDir string) *DownloadItem {
	item := downloadManager.Start(rawURL)

	active := &ActiveDownload{Item: item}
	activeDownloads.mu.Lock()
	activeDownloads.Items[item.ID] = active
	activeDownloads.mu.Unlock()

	go func() {
		err := doDownload(rawURL, saveDir, item, active)
		if err != nil {
			downloadManager.Fail(item.ID)
			return
		}
		downloadManager.Complete(item.ID)
	}()

	return item
}

func doDownload(rawURL string, saveDir string, item *DownloadItem, active *ActiveDownload) error {
	resp, err := http.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errNotOK(resp.StatusCode)
	}

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return err
	}

	fileName := sanitizeFileName(item.FileName)
	fullPath := filepath.Join(saveDir, fileName)

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)

	for {
		active.mu.Lock()
		cancelled := active.Cancelled
		active.mu.Unlock()

		if cancelled {
			downloadManager.Cancel(item.ID)
			os.Remove(fullPath)
			return nil
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			written += int64(n)
			if total > 0 {
				percent := int((written * 100) / total)
				downloadManager.UpdateProgress(item.ID, percent)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	downloadManager.UpdateProgress(item.ID, 100)
	return nil
}

// CancelActiveDownload signals the goroutine to stop and delete partial file
func CancelActiveDownload(id int) {
	activeDownloads.mu.Lock()
	active, ok := activeDownloads.Items[id]
	activeDownloads.mu.Unlock()

	if !ok {
		return
	}
	active.mu.Lock()
	active.Cancelled = true
	active.mu.Unlock()
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "download"
	}
	replacer := strings.NewReplacer(
		"<", "_", ">", "_", ":", "_", "\"", "_",
		"/", "_", "\\", "_", "|", "_", "?", "_", "*", "_",
	)
	return replacer.Replace(name)
}

type downloadHTTPError struct {
	code int
}

func (e *downloadHTTPError) Error() string {
	return "download failed with status code"
}

func errNotOK(code int) error {
	return &downloadHTTPError{code: code}
}

// IsDownloadableURL checks file extensions that should trigger a download
// instead of opening in the browser (basic heuristic)
func IsDownloadableURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	downloadExts := []string{
		".zip", ".rar", ".7z", ".exe", ".msi", ".dmg",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".mp3", ".mp4", ".mkv", ".avi", ".mov",
		".iso", ".apk", ".tar", ".gz",
	}
	for _, ext := range downloadExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}