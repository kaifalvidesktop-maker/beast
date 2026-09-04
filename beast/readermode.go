package main

import "sync"

// ---------------------------------------------------
// READER MODE
// Tracks which tabs currently have reader (distraction-free) mode on
// Actual content stripping happens in injected JS on the page side
// ---------------------------------------------------

type ReaderModeManager struct {
	mu     sync.Mutex
	Active map[int]bool // tabID -> is reader mode on
}

var readerMode = &ReaderModeManager{
	Active: make(map[int]bool),
}

// Toggle reader mode for a tab
func (r *ReaderModeManager) Toggle(tabID int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Active[tabID] = !r.Active[tabID]
	return r.Active[tabID]
}

// Check if reader mode is on for a tab
func (r *ReaderModeManager) IsActive(tabID int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Active[tabID]
}

// Turn off reader mode when navigating away
func (r *ReaderModeManager) Clear(tabID int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Active, tabID)
}

// Injectable JS that strips clutter and shows clean article view
const readerModeJS = `
(function() {
	if (document.getElementById('beast-reader-style')) return;
	var style = document.createElement('style');
	style.id = 'beast-reader-style';
	style.innerHTML =
		'body > *:not(article):not(main) { display: none !important; }' +
		'article, main { max-width: 700px !important; margin: 40px auto !important; ' +
		'font-size: 19px !important; line-height: 1.7 !important; ' +
		'background: #16171a !important; color: #e8e8e8 !important; padding: 40px !important; }';
	document.head.appendChild(style);
})();
`