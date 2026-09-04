package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ---------------------------------------------------
// FULL DATA BACKUP / RESTORE
// Bundles bookmarks, settings, and shortcuts into one JSON
// file the user can save and restore later — this is how
// "no account, no sync" still lets users move between machines.
// Note: passwords are intentionally excluded from plain backup;
// see ExportEncryptedVault() below for those.
// ---------------------------------------------------

type BackupBundle struct {
	Version   string             `json:"version"`
	CreatedAt string             `json:"createdAt"`
	Bookmarks []ExportedBookmark `json:"bookmarks"`
	Settings  map[string]any     `json:"settings"`
	Shortcuts []PinnedShortcut   `json:"shortcuts"`
}

// BuildBackup assembles all exportable data into one bundle
func BuildBackup() BackupBundle {
	all := bookmarkManager.GetAll()
	exportList := make([]ExportedBookmark, 0, len(all))
	for _, b := range all {
		exportList = append(exportList, ExportedBookmark{
			Title: b.Title, URL: b.URL, Folder: b.Folder,
		})
	}

	return BackupBundle{
		Version:   beastVersion,
		CreatedAt: time.Now().Format(time.RFC3339),
		Bookmarks: exportList,
		Settings:  settings.GetAll(),
		Shortcuts: shortcutManagerNewTab.GetShortcuts(),
	}
}

// ExportBackupJSON returns the full backup as a JSON string
func ExportBackupJSON() string {
	bundle := BuildBackup()
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// SaveBackupToFile writes the backup JSON to the downloads folder
func SaveBackupToFile() (string, error) {
	json := ExportBackupJSON()

	if err := os.MkdirAll(settings.DownloadPath, 0755); err != nil {
		return "", err
	}

	fileName := "beast-backup_" + time.Now().Format("2006-01-02_15-04-05") + ".json"
	fullPath := filepath.Join(settings.DownloadPath, fileName)

	err := os.WriteFile(fullPath, []byte(json), 0644)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

// RestoreBackupJSON parses a backup JSON string and applies it
func RestoreBackupJSON(jsonStr string) (int, error) {
	var bundle BackupBundle
	if err := json.Unmarshal([]byte(jsonStr), &bundle); err != nil {
		return 0, err
	}

	restoredCount := 0

	// Restore bookmarks (skip duplicates)
	for _, b := range bundle.Bookmarks {
		if b.URL == "" {
			continue
		}
		if !bookmarkManager.IsBookmarked(b.URL) {
			bookmarkManager.Add(b.Title, b.URL, b.Folder)
			restoredCount++
		}
	}

	// Restore settings (best-effort, only known fields)
	if theme, ok := bundle.Settings["theme"].(string); ok {
		settings.SetTheme(theme)
	}
	if engine, ok := bundle.Settings["searchEngine"].(string); ok {
		settings.SetSearchEngine(engine)
	}
	if homepage, ok := bundle.Settings["homepage"].(string); ok {
		settings.SetHomepage(homepage)
	}

	// Restore shortcuts
	for _, s := range bundle.Shortcuts {
		shortcutManagerNewTab.AddShortcut(s.Title, s.URL)
	}

	return restoredCount, nil
}
