package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"os"
	"path/filepath"
	"plugin"
	"sync"
	"time"
)

type ExtensionManifest struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	Description    string   `json:"description"`
	Author         string   `json:"author"`
	Permissions    []string `json:"permissions"`
	Background     string   `json:"background"`
	ContentScripts []string `json:"content_scripts"`
	Popup          string   `json:"popup"`
	Icon           string   `json:"icon"`
}

type Extension struct {
	Manifest ExtensionManifest
	Path     string
	Enabled  bool
	LoadedAt time.Time
	Plugin   *plugin.Plugin
}

type ExtensionManager struct {
	mu         sync.RWMutex
	extensions map[string]*Extension
	extDir     string
}

// NewExtensionManager creates a new extension manager
func NewExtensionManager(extDir string) *ExtensionManager {
	os.MkdirAll(extDir, 0755)
	return &ExtensionManager{
		extensions: make(map[string]*Extension),
		extDir:     extDir,
	}
}

// InstallExtension installs an extension from a directory
func (em *ExtensionManager) InstallExtension(path string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Read manifest
	manifestPath := filepath.Join(path, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ExtensionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	if manifest.ID == "" {
		return errors.New("extension ID is required")
	}

	// Check if already installed
	if _, exists := em.extensions[manifest.ID]; exists {
		return errors.New("extension already installed")
	}

	ext := &Extension{
		Manifest: manifest,
		Path:     path,
		Enabled:  true,
		LoadedAt: time.Now(),
	}

	// Try to load as Go plugin
	pluginPath := filepath.Join(path, "extension.so")
	if _, err := os.Stat(pluginPath); err == nil {
		if p, err := plugin.Open(pluginPath); err == nil {
			ext.Plugin = p
		}
	}

	em.extensions[manifest.ID] = ext
	return nil
}

// EnableExtension enables a disabled extension
func (em *ExtensionManager) EnableExtension(id string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if ext, exists := em.extensions[id]; exists {
		ext.Enabled = true
		return nil
	}
	return errors.New("extension not found")
}

// DisableExtension disables an enabled extension
func (em *ExtensionManager) DisableExtension(id string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if ext, exists := em.extensions[id]; exists {
		ext.Enabled = false
		return nil
	}
	return errors.New("extension not found")
}

// UninstallExtension removes an extension
func (em *ExtensionManager) UninstallExtension(id string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.extensions[id]; exists {
		delete(em.extensions, id)
		return nil
	}
	return errors.New("extension not found")
}

// GetExtension returns an extension by ID
func (em *ExtensionManager) GetExtension(id string) *Extension {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.extensions[id]
}

// GetAllExtensions returns all extensions
func (em *ExtensionManager) GetAllExtensions() []*Extension {
	em.mu.RLock()
	defer em.mu.RUnlock()

	exts := make([]*Extension, 0, len(em.extensions))
	for _, ext := range em.extensions {
		exts = append(exts, ext)
	}
	return exts
}

// GetEnabledExtensions returns enabled extensions
func (em *ExtensionManager) GetEnabledExtensions() []*Extension {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var enabled []*Extension
	for _, ext := range em.extensions {
		if ext.Enabled {
			enabled = append(enabled, ext)
		}
	}
	return enabled
}

// GenerateExtensionsPageHTML generates HTML for extensions page
func (em *ExtensionManager) GenerateExtensionsPageHTML() string {
	extensions := em.GetAllExtensions()

	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Extensions - Beast Browser</title>
    <style>
        body { background: #1a1a2e; color: #eee; font-family: 'Segoe UI', Arial, sans-serif; padding: 20px; }
        h1 { color: #e94560; border-bottom: 2px solid #e94560; padding-bottom: 10px; }
        .ext-card {
            background: #16213e;
            padding: 20px;
            margin: 15px 0;
            border-radius: 8px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .ext-info h3 { color: #6bcb77; margin: 0; }
        .ext-info p { color: #888; margin: 5px 0; }
        .ext-version { color: #666; font-size: 0.9em; }
        .ext-badge {
            background: #333;
            padding: 2px 10px;
            border-radius: 12px;
            font-size: 0.8em;
        }
        .ext-enabled { background: #6bcb77; color: #1a1a2e; }
        .ext-disabled { background: #dc3545; color: white; }
        .btn {
            background: #e94560;
            color: white;
            border: none;
            padding: 5px 15px;
            border-radius: 5px;
            cursor: pointer;
            margin: 0 5px;
        }
        .btn:hover { background: #c73652; }
        .btn-enable { background: #6bcb77; }
        .btn-enable:hover { background: #52b95e; }
        .btn-disable { background: #ffa94d; }
        .btn-disable:hover { background: #f08c3a; }
        .btn-danger { background: #dc3545; }
        .btn-danger:hover { background: #c82333; }
        .install-section {
            background: #16213e;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .install-section input {
            padding: 10px;
            border-radius: 5px;
            border: 1px solid #333;
            background: #1a1a2e;
            color: #eee;
            width: 70%;
            margin-right: 10px;
        }
    </style>
</head>
<body>
    <h1>🧩 Extensions</h1>
    
    <div class="install-section">
        <input type="text" id="extPath" placeholder="Extension folder path...">
        <button class="btn" onclick="installExtension()">📦 Install</button>
    </div>`)

	if len(extensions) == 0 {
		html.WriteString(`<p>No extensions installed.</p>`)
	} else {
		for _, ext := range extensions {
			statusClass := "ext-disabled"
			statusText := "Disabled"
			if ext.Enabled {
				statusClass = "ext-enabled"
				statusText = "Enabled"
			}

			html.WriteString(fmt.Sprintf(`
    <div class="ext-card">
        <div class="ext-info">
            <h3>%s</h3>
            <p>%s</p>
            <div>
                <span class="ext-version">v%s</span>
                <span class="ext-badge">%s</span>
                <span class="ext-badge %s">%s</span>
                <span style="color:#888;font-size:0.8em;margin-left:10px;">%s</span>
            </div>
        </div>
        <div>
            <button class="btn %s" onclick="toggleExtension('%s')">%s</button>
            <button class="btn btn-danger" onclick="uninstallExtension('%s')">🗑️ Remove</button>
        </div>
    </div>`,
				ext.Manifest.Name, ext.Manifest.Description,
				ext.Manifest.Version, ext.Manifest.Author,
				statusClass, statusText, ext.LoadedAt.Format("Jan 2, 15:04"),
				func() string {
					if ext.Enabled {
						return "btn-disable"
					}
					return "btn-enable"
				}(),
				ext.Manifest.ID,
				func() string {
					if ext.Enabled {
						return "Disable"
					}
					return "Enable"
				}(),
				ext.Manifest.ID))
		}
	}

	html.WriteString(`
    <script>
        function installExtension() {
            const path = document.getElementById('extPath').value;
            if(path) {
                window.location.href = 'beast://installextension?path=' + encodeURIComponent(path);
            }
        }
        function toggleExtension(id) {
            window.location.href = 'beast://toggleextension?id=' + id;
        }
        function uninstallExtension(id) {
            if(confirm('Remove this extension?')) {
                window.location.href = 'beast://uninstallextension?id=' + id;
            }
        }
    </script>
</body>
</html>`)
	return html.String()
}
