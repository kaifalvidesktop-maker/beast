package main

import (
	"strings"
)

// ---------------------------------------------------
// REAL AD / TRACKER BLOCKING
// ---------------------------------------------------
//
// The old "shield" in security.go only checked the URL typed into the
// address bar - it never touched what a page itself loads (ad scripts,
// tracking pixels, analytics beacons, third-party iframes). That meant
// visiting any normal website bypassed it completely.
//
// webview_go doesn't expose a native "block this network request" hook
// on every platform, so instead this blocks at the JS layer: BuildAdblockJS
// returns a script that webview.Init() runs before any page script,
// on every navigation. It:
//
//   1. Patches window.fetch and XMLHttpRequest so requests to a blocked
//      domain are rejected/aborted before they leave the page.
//   2. Watches the DOM (MutationObserver) and removes/blanks <script>,
//      <img>, <iframe>, and <link> tags whose src/href hosts match a
//      blocked domain, including ones inserted after page load.
//   3. Reports each block back to Go via the "reportBlocked" binding so
//      shield.BlockedCount stays accurate for the UI.
//
// The block lists themselves stay defined once, in security.go
// (adDomains / trackerDomains), and are serialized into the script here
// so there's a single source of truth.

// buildDomainListJS turns a Go domain set into a JS array literal.
func buildDomainListJS(domains map[string]bool) string {
	names := make([]string, 0, len(domains))
	for d := range domains {
		names = append(names, `"`+d+`"`)
	}
	return "[" + strings.Join(names, ",") + "]"
}

// BuildAdblockJS returns the full injectable script. It's regenerated
// each call so it always reflects the live enabled/disabled state and
// domain lists from shield.
func BuildAdblockJS() string {
	shield.mu.Lock()
	adOn := shield.AdBlockOn
	trackerOn := shield.TrackerBlockOn
	shield.mu.Unlock()

	adList := buildDomainListJS(adDomains)
	trackerList := buildDomainListJS(trackerDomains)

	return `(function(){
	"use strict";
	var AD_ON = ` + boolJS(adOn) + `;
	var TRACKER_ON = ` + boolJS(trackerOn) + `;
	var AD_DOMAINS = ` + adList + `;
	var TRACKER_DOMAINS = ` + trackerList + `;

	function hostFromURL(u) {
		try { return new URL(u, window.location.href).hostname.toLowerCase(); }
		catch (e) { return ""; }
	}

	function matchesList(host, list) {
		if (!host) return false;
		for (var i = 0; i < list.length; i++) {
			var d = list[i];
			if (host === d || host.endsWith("." + d)) return true;
		}
		return false;
	}

	function isBlocked(url) {
		var host = hostFromURL(url);
		if (AD_ON && matchesList(host, AD_DOMAINS)) return "ad";
		if (TRACKER_ON && matchesList(host, TRACKER_DOMAINS)) return "tracker";
		return null;
	}

	function reportBlock(kind, url) {
		if (window.reportBlocked) {
			try { window.reportBlocked(kind, url); } catch (e) {}
		}
	}

	// --- fetch() ---
	var nativeFetch = window.fetch;
	if (nativeFetch) {
		window.fetch = function(input, init) {
			var url = (typeof input === "string") ? input : (input && input.url);
			var kind = isBlocked(url || "");
			if (kind) {
				reportBlock(kind, url);
				return Promise.reject(new TypeError("BEAST Shield blocked: " + url));
			}
			return nativeFetch.apply(this, arguments);
		};
	}

	// --- XMLHttpRequest ---
	var nativeOpen = XMLHttpRequest.prototype.open;
	XMLHttpRequest.prototype.open = function(method, url) {
		var kind = isBlocked(url || "");
		if (kind) {
			reportBlock(kind, url);
			this.__beastBlocked = true;
			// Point at a data: URL instead of the real request so nothing fires.
			arguments[1] = "data:text/plain,";
		}
		return nativeOpen.apply(this, arguments);
	};

	// --- DOM elements: script / img / iframe / link ---
	function checkAndStrip(el) {
		if (!el || !el.tagName) return;
		var tag = el.tagName.toLowerCase();
		var attr = (tag === "link") ? "href" : "src";
		var url = el.getAttribute && el.getAttribute(attr);
		if (!url) return;
		var kind = isBlocked(url);
		if (kind && (tag === "script" || tag === "img" || tag === "iframe" || tag === "link")) {
			reportBlock(kind, url);
			el.removeAttribute(attr);
			if (el.parentNode) el.parentNode.removeChild(el);
		}
	}

	function scanExisting() {
		document.querySelectorAll("script[src], img[src], iframe[src], link[href]")
			.forEach(checkAndStrip);
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", scanExisting);
	} else {
		scanExisting();
	}

	var observer = new MutationObserver(function(mutations) {
		mutations.forEach(function(m) {
			m.addedNodes && m.addedNodes.forEach(function(node) {
				checkAndStrip(node);
				if (node.querySelectorAll) {
					node.querySelectorAll("script[src], img[src], iframe[src], link[href]")
						.forEach(checkAndStrip);
				}
			});
		});
	});

	function startObserving() {
		observer.observe(document.documentElement || document, {
			childList: true, subtree: true
		});
	}

	if (document.documentElement) {
		startObserving();
	} else {
		document.addEventListener("DOMContentLoaded", startObserving);
	}
})();`
}

func boolJS(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
