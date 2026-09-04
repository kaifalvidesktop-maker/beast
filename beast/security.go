package main

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// ---------------------------------------------------
// SECURITY SHIELD SYSTEM
// ---------------------------------------------------

type SecurityShield struct {
	mu             sync.Mutex
	AdBlockOn      bool
	TrackerBlockOn bool
	HTTPSOnlyOn    bool
	BlockedCount   int
}

var shield = &SecurityShield{
	AdBlockOn:      true,
	TrackerBlockOn: true,
	HTTPSOnlyOn:    true,
	BlockedCount:   0,
}

// Known Ad Domains (basic starter list, will grow later)
var adDomains = map[string]bool{
	"doubleclick.net":       true,
	"googlesyndication.com": true,
	"googleadservices.com":  true,
	"adservice.google.com":  true,
	"ads.yahoo.com":         true,
	"adnxs.com":             true,
	"popads.net":            true,
	"outbrain.com":          true,
	"taboola.com":           true,
	"advertising.com":       true,
}

// Known Tracker Domains
var trackerDomains = map[string]bool{
	"google-analytics.com":  true,
	"googletagmanager.com":  true,
	"facebook.net":          true,
	"connect.facebook.net":  true,
	"hotjar.com":            true,
	"segment.io":            true,
	"mixpanel.com":          true,
	"amplitude.com":         true,
	"scorecardresearch.com": true,
	"quantserve.com":        true,
}

// Extract domain from a URL string
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	return strings.ToLower(host)
}

// Check if a domain (or its parent domain) is in a block list
func isDomainBlocked(domain string, list map[string]bool) bool {
	if domain == "" {
		return false
	}
	if list[domain] {
		return true
	}
	// Check parent domains too (e.g. sub.doubleclick.net)
	parts := strings.Split(domain, ".")
	for i := 0; i < len(parts)-1; i++ {
		sub := strings.Join(parts[i:], ".")
		if list[sub] {
			return true
		}
	}
	return false
}

// Main function: should this request be blocked?
func (s *SecurityShield) ShouldBlock(rawURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	domain := extractDomain(rawURL)

	if s.AdBlockOn && isDomainBlocked(domain, adDomains) {
		s.BlockedCount++
		fmt.Println("[AD BLOCKED]", domain)
		return true
	}

	if s.TrackerBlockOn && isDomainBlocked(domain, trackerDomains) {
		s.BlockedCount++
		fmt.Println("[TRACKER BLOCKED]", domain)
		return true
	}

	return false
}

// Force HTTPS on a URL if HTTPS-Only mode is on
func (s *SecurityShield) EnforceHTTPS(rawURL string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.HTTPSOnlyOn {
		return rawURL
	}

	if strings.HasPrefix(rawURL, "http://") {
		return strings.Replace(rawURL, "http://", "https://", 1)
	}

	return rawURL
}

// Toggle functions for UI switches
func (s *SecurityShield) ToggleAdBlock() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AdBlockOn = !s.AdBlockOn
	return s.AdBlockOn
}

func (s *SecurityShield) ToggleTrackerBlock() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TrackerBlockOn = !s.TrackerBlockOn
	return s.TrackerBlockOn
}

func (s *SecurityShield) ToggleHTTPSOnly() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.HTTPSOnlyOn = !s.HTTPSOnlyOn
	return s.HTTPSOnlyOn
}

// Get current shield stats (for UI display)
func (s *SecurityShield) GetStats() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"adBlock":      s.AdBlockOn,
		"trackerBlock": s.TrackerBlockOn,
		"httpsOnly":    s.HTTPSOnlyOn,
		"blockedCount": s.BlockedCount,
	}
}

// ---------------------------------------------------
// URL VALIDATION / SEARCH LOGIC
// ---------------------------------------------------

// Decide if user typed a URL or a search query
func resolveInput(input string) string {
	input = strings.TrimSpace(input)

	if input == "" {
		return ""
	}

	// If it already looks like a URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return input
	}

	// If it looks like a domain (has a dot, no spaces)
	if strings.Contains(input, ".") && !strings.Contains(input, " ") {
		return "https://" + input
	}

	// Otherwise treat as a Google search
	query := strings.ReplaceAll(input, " ", "+")
	return "https://www.google.com/search?q=" + query
}
