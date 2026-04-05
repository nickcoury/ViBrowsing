package window

import (
	"runtime"
	"sync"
	"time"
)

// Navigator represents the navigator object accessible via JavaScript.
// It provides information about the browser's properties.
type Navigator struct {
	UserAgent string
	// Connection info (NetworkInformation API)
	connection *NetworkInformation
	mu         sync.RWMutex
}

// NetworkInformation represents the NetworkInformation API (navigator.connection).
// This provides information about the device's network connection.
type NetworkInformation struct {
	// EffectiveConnectionType returns the type of connection:
	// "slow-2g", "2g", "3g", "4g", or "unknown"
	EffectiveConnectionType string
	// Downlink returns the effective bandwidth estimate in Mbps (0 if unknown)
	Downlink float64
	// RTT returns the round-trip time estimate in milliseconds (0 if unknown)
	RTT int
	// SaveData returns true if the user has enabled data savings mode
	SaveData bool
	// Type returns the type of connection: "wifi", "cellular", "ethernet", "unknown"
	Type string
	// OnChange is called when connection properties change
	OnChange func()
}

// NewNavigator creates a new Navigator with the default user agent.
func NewNavigator() *Navigator {
	n := &Navigator{
		UserAgent: "HallucinHTML/1.0 (+https://github.com/nickcoury/ViBrowsing)",
	}
	n.connection = n.detectConnection()
	return n
}

// NewNavigatorWithAgent creates a new Navigator with a custom user agent.
func NewNavigatorWithAgent(userAgent string) *Navigator {
	ua := userAgent
	if ua == "" {
		ua = "HallucinHTML/1.0 (+https://github.com/nickcoury/ViBrowsing)"
	}
	n := &Navigator{
		UserAgent: ua,
	}
	n.connection = n.detectConnection()
	return n
}

// detectConnection detects the network connection characteristics.
func (n *Navigator) detectConnection() *NetworkInformation {
	ni := &NetworkInformation{
		EffectiveConnectionType: "unknown",
		Downlink:                0,
		RTT:                     0,
		SaveData:                false,
		Type:                    "unknown",
	}

	// Try to detect connection type based on platform and available info
	// In a real browser, this would use the Network Information API
	// For ViBrowsing, we estimate based on platform

	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		// Assume broadband connection on desktop platforms
		ni.EffectiveConnectionType = "4g"
		ni.Downlink = 10.0 // Conservative estimate in Mbps
		ni.RTT = 20        // 20ms round-trip time
		ni.Type = "ethernet"
	default:
		// Mobile or unknown
		ni.EffectiveConnectionType = "3g"
		ni.Downlink = 1.5
		ni.RTT = 100
		ni.Type = "cellular"
	}

	return ni
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

// Connection returns the NetworkInformation object (navigator.connection).
func (n *Navigator) Connection() *NetworkInformation {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.connection
}

// SetSaveData sets the save data mode (data saver).
func (n *Navigator) SetSaveData(enabled bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.connection != nil {
		n.connection.SaveData = enabled
	}
}

// SetEffectiveType sets the effective connection type.
func (n *Navigator) SetEffectiveType(connType string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.connection != nil {
		n.connection.EffectiveConnectionType = connType
	}
}

// SetDownlink sets the estimated bandwidth in Mbps.
func (n *Navigator) SetDownlink(mbps float64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.connection != nil {
		n.connection.Downlink = mbps
	}
}

// SetRTT sets the round-trip time in milliseconds.
func (n *Navigator) SetRTT(ms int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.connection != nil {
		n.connection.RTT = ms
	}
}

// Language returns the primary language of the browser.
func (n *Navigator) Language() string {
	return "en-US"
}

// Languages returns the list of preferred languages.
func (n *Navigator) Languages() []string {
	return []string{"en-US", "en"}
}

// HardwareConcurrency returns the number of logical CPU cores.
func (n *Navigator) HardwareConcurrency() int {
	return runtime.NumCPU()
}

// DeviceMemory returns the approximate amount of device memory in GB.
// Returns 0 if unknown.
func (n *Navigator) DeviceMemory() float64 {
	// Go doesn't have a direct way to get device memory
	// Return a reasonable default
	var memGB float64 = 8
	return memGB
}

// MaxTouchPoints returns the maximum number of simultaneous touch points.
func (n *Navigator) MaxTouchPoints() int {
	return 0 // Non-touch device by default
}

// CookieEnabled returns true if cookies are enabled.
func (n *Navigator) CookieEnabled() bool {
	return true
}

// JavaEnabled returns true if Java is enabled (always false for ViBrowsing).
func (n *Navigator) JavaEnabled() bool {
	return false
}

// Online returns true if the browser is online.
func (n *Navigator) Online() bool {
	return true
}

// DoNotTrack returns the Do Not Track setting.
func (n *Navigator) DoNotTrack() string {
	return "1"
}

// Webdriver returns true if the browser is controlled by WebDriver.
func (n *Navigator) Webdriver() bool {
	return false
}

// Product returns the product name (for compatibility).
func (n *Navigator) Product() string {
	return "Gecko"
}

// ProductSub returns the product sub name (for compatibility).
func (n *Navigator) ProductSub() string {
	return "20030107"
}

// Vendor returns the vendor name (for compatibility).
func (n *Navigator) Vendor() string {
	return "Google"
}

// VendorSub returns the vendor sub name (for compatibility).
func (n *Navigator) VendorSub() string {
	return ""
}

// UpdateConnectionForTest updates connection info for testing purposes.
func (n *Navigator) UpdateConnectionForTest(connType string, downlink float64, rtt int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.connection != nil {
		n.connection.EffectiveConnectionType = connType
		n.connection.Downlink = downlink
		n.connection.RTT = rtt
	}
}

// ConnectionChangeNotifier allows observing connection changes.
type ConnectionChangeNotifier struct {
	mu      sync.RWMutex
	handler func(*NetworkInformation)
}

// OnConnectionChange sets a handler for connection changes.
func (n *Navigator) OnConnectionChange(handler func()) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.connection != nil {
		n.connection.OnChange = handler
	}
}

// SimulateConnectionChange simulates a connection change event.
func (n *Navigator) SimulateConnectionChange(connType string, downlink float64, rtt int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.connection != nil {
		n.connection.EffectiveConnectionType = connType
		n.connection.Downlink = downlink
		n.connection.RTT = rtt
		if n.connection.OnChange != nil {
			go n.connection.OnChange()
		}
	}
}

// Last fetch timing information
type FetchTiming struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}
