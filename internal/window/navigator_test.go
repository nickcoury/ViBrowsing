package window

import (
	"sync"
	"testing"
	"time"
)

func TestNavigatorConnection(t *testing.T) {
	n := NewNavigator()
	
	conn := n.Connection()
	if conn == nil {
		t.Fatal("Connection() returned nil")
	}

	// Check default values are set
	if conn.EffectiveConnectionType == "" {
		t.Error("EffectiveConnectionType should not be empty")
	}
	if conn.Downlink < 0 {
		t.Error("Downlink should be non-negative")
	}
	if conn.RTT < 0 {
		t.Error("RTT should be non-negative")
	}
}

func TestNavigatorConnectionTypes(t *testing.T) {
	n := NewNavigator()
	
	// Test setting different connection types
	tests := []struct {
		connType string
		downlink float64
		rtt      int
	}{
		{"slow-2g", 0.05, 2000},
		{"2g", 0.15, 600},
		{"3g", 0.4, 200},
		{"4g", 10.0, 20},
	}

	for _, tt := range tests {
		t.Run(tt.connType, func(t *testing.T) {
			n.UpdateConnectionForTest(tt.connType, tt.downlink, tt.rtt)
			conn := n.Connection()
			
			if conn.EffectiveConnectionType != tt.connType {
				t.Errorf("EffectiveConnectionType = %q, want %q", conn.EffectiveConnectionType, tt.connType)
			}
			if conn.Downlink != tt.downlink {
				t.Errorf("Downlink = %v, want %v", conn.Downlink, tt.downlink)
			}
			if conn.RTT != tt.rtt {
				t.Errorf("RTT = %d, want %d", conn.RTT, tt.rtt)
			}
		})
	}
}

func TestNavigatorSaveData(t *testing.T) {
	n := NewNavigator()
	
	conn := n.Connection()
	if conn.SaveData {
		t.Error("SaveData should be false by default")
	}

	n.SetSaveData(true)
	conn = n.Connection()
	if !conn.SaveData {
		t.Error("SaveData should be true after SetSaveData(true)")
	}

	n.SetSaveData(false)
	conn = n.Connection()
	if conn.SaveData {
		t.Error("SaveData should be false after SetSaveData(false)")
	}
}

func TestNavigatorConnectionOnChange(t *testing.T) {
	n := NewNavigator()
	
	var changeCount int
	var mu sync.Mutex
	n.OnConnectionChange(func() {
		mu.Lock()
		changeCount++
		mu.Unlock()
	})

	// Simulate a connection change
	n.SimulateConnectionChange("3g", 1.5, 100)
	
	// Give goroutine time to execute
	time.Sleep(10 * time.Millisecond)
	
	// Handler should have been called
	mu.Lock()
	defer mu.Unlock()
	if changeCount != 1 {
		t.Errorf("OnChange handler called %d times, want 1", changeCount)
	}
}

func TestNavigatorBasicProperties(t *testing.T) {
	n := NewNavigator()

	tests := []struct {
		method string
		got    string
	}{
		{"AppCodeName", n.AppCodeName()},
		{"AppName", n.AppName()},
		{"AppVersion", n.AppVersion()},
		{"Platform", n.Platform()},
		{"Language", n.Language()},
		{"Product", n.Product()},
		{"Vendor", n.Vendor()},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			if tt.got == "" {
				t.Errorf("%s() returned empty string", tt.method)
			}
		})
	}
}

func TestNavigatorLanguages(t *testing.T) {
	n := NewNavigator()
	
	languages := n.Languages()
	if len(languages) == 0 {
		t.Error("Languages() returned empty slice")
	}
	
	if languages[0] != "en-US" {
		t.Errorf("First language = %q, want en-US", languages[0])
	}
}

func TestNavigatorHardwareConcurrency(t *testing.T) {
	n := NewNavigator()
	
	concurrency := n.HardwareConcurrency()
	if concurrency <= 0 {
		t.Error("HardwareConcurrency() should be positive")
	}
}

func TestNavigatorCookieEnabled(t *testing.T) {
	n := NewNavigator()
	
	if !n.CookieEnabled() {
		t.Error("CookieEnabled() should return true")
	}
}

func TestNavigatorJavaEnabled(t *testing.T) {
	n := NewNavigator()
	
	if n.JavaEnabled() {
		t.Error("JavaEnabled() should return false")
	}
}

func TestNavigatorOnline(t *testing.T) {
	n := NewNavigator()
	
	if !n.Online() {
		t.Error("Online() should return true")
	}
}

func TestNavigatorWebdriver(t *testing.T) {
	n := NewNavigator()
	
	if n.Webdriver() {
		t.Error("Webdriver() should return false")
	}
}

func TestNavigatorDeviceMemory(t *testing.T) {
	n := NewNavigator()
	
	mem := n.DeviceMemory()
	if mem <= 0 {
		t.Error("DeviceMemory() should be positive")
	}
}

func TestNavigatorMaxTouchPoints(t *testing.T) {
	n := NewNavigator()
	
	points := n.MaxTouchPoints()
	if points < 0 {
		t.Error("MaxTouchPoints() should be non-negative")
	}
}

func TestNewNavigatorWithAgent(t *testing.T) {
	customUA := "CustomAgent/1.0"
	n := NewNavigatorWithAgent(customUA)
	
	if n.UserAgent != customUA {
		t.Errorf("UserAgent = %q, want %q", n.UserAgent, customUA)
	}
}

func TestNewNavigatorWithEmptyAgent(t *testing.T) {
	n := NewNavigatorWithAgent("")
	
	if n.UserAgent == "" {
		t.Error("UserAgent should not be empty with empty input")
	}
}

func TestConnectionTypeValues(t *testing.T) {
	n := NewNavigator()
	
	// All valid connection types should be accepted
	validTypes := []string{"slow-2g", "2g", "3g", "4g", "unknown"}
	
	for _, ct := range validTypes {
		n.SetEffectiveType(ct)
		conn := n.Connection()
		if conn.EffectiveConnectionType != ct {
			t.Errorf("SetEffectiveType(%q) didn't stick, got %q", ct, conn.EffectiveConnectionType)
		}
	}
}

func TestDownlinkValues(t *testing.T) {
	n := NewNavigator()
	
	tests := []float64{0.0, 0.5, 1.0, 5.0, 10.0, 100.0}
	
	for _, dl := range tests {
		n.SetDownlink(dl)
		conn := n.Connection()
		if conn.Downlink != dl {
			t.Errorf("SetDownlink(%v) didn't stick, got %v", dl, conn.Downlink)
		}
	}
}

func TestRTTValues(t *testing.T) {
	n := NewNavigator()
	
	tests := []int{0, 50, 100, 200, 500, 1000}
	
	for _, rtt := range tests {
		n.SetRTT(rtt)
		conn := n.Connection()
		if conn.RTT != rtt {
			t.Errorf("SetRTT(%d) didn't stick, got %d", rtt, conn.RTT)
		}
	}
}
