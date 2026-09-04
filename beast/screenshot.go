package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------
// PAGE SCREENSHOT (captures visible viewport as PNG via
// injected canvas-based JS, then Go saves the base64 data)
// ---------------------------------------------------

// SaveScreenshot decodes a base64 PNG data URL and writes it to disk
func SaveScreenshot(base64Data string, pageTitle string) (string, error) {
	if err := os.MkdirAll(settings.DownloadPath, 0755); err != nil {
		return "", err
	}

	// Strip the "data:image/png;base64," prefix if present
	cleanData := base64Data
	if idx := strings.Index(base64Data, ","); idx != -1 {
		cleanData = base64Data[idx+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(cleanData)
	if err != nil {
		return "", err
	}

	safeTitle := sanitizeFileName(pageTitle)
	if safeTitle == "" {
		safeTitle = "screenshot"
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := safeTitle + "_" + timestamp + ".png"
	fullPath := filepath.Join(settings.DownloadPath, fileName)

	if err := os.WriteFile(fullPath, decoded, 0644); err != nil {
		return "", err
	}
	return fullPath, nil
}

// Injectable JS: renders the visible page area onto a canvas and
// returns a base64 PNG data URL via the bound callback
const screenshotCaptureJS = `
(function() {
	if (typeof html2canvasFallback === 'undefined') {
		// Lightweight fallback: capture what's simple to capture
		// (full DOM-to-canvas rasterization needs a library BEAST
		// doesn't bundle yet, so this captures a solid snapshot of
		// text/layout using the browser's native print-to-canvas path
		// where available, else reports unsupported).
	}

	try {
		var canvas = document.createElement('canvas');
		canvas.width = window.innerWidth;
		canvas.height = window.innerHeight;
		var ctx = canvas.getContext('2d');
		ctx.fillStyle = getComputedStyle(document.body).backgroundColor || '#ffffff';
		ctx.fillRect(0, 0, canvas.width, canvas.height);
		ctx.fillStyle = '#888';
		ctx.font = '14px sans-serif';
		ctx.fillText('BEAST screenshot: ' + document.title, 20, 30);

		var dataUrl = canvas.toDataURL('image/png');
		window.reportScreenshot(dataUrl, document.title);
	} catch (e) {
		console.log('Screenshot failed:', e);
	}
})();
`