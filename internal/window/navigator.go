package window

// Navigator represents the navigator object accessible via JavaScript.
// It provides information about the browser's properties.
type Navigator struct {
	// UserAgent is the user agent string sent to servers
	UserAgent string
}

// NewNavigator creates a new Navigator with the default user agent.
func NewNavigator() *Navigator {
	return &Navigator{
		UserAgent: "HallucinHTML/1.0 (+https://github.com/nickcoury/ViBrowsing)",
	}
}

// NewNavigatorWithAgent creates a new Navigator with a custom user agent.
func NewNavigatorWithAgent(userAgent string) *Navigator {
	ua := userAgent
	if ua == "" {
		ua = "HallucinHTML/1.0 (+https://github.com/nickcoury/ViBrowsing)"
	}
	return &Navigator{
		UserAgent: ua,
	}
}

// AppCodeName returns the code name of the browser (for compatibility).
func (n *Navigator) AppCodeName() string {
	return "Mozilla"
}

// AppName returns the name of the browser (for compatibility).
func (n *Navigator) AppName() string {
	return "Netscape"
}

// AppVersion returns the version of the browser (for compatibility).
func (n *Navigator) AppVersion() string {
	return "5.0 (X11)"
}

// Platform returns the platform the browser is running on.
func (n *Navigator) Platform() string {
	return "Linux x86_64"
}
