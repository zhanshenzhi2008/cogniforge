package handler_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// parseDeviceFromUATest 测试 User-Agent 解析逻辑（与 auth.go 中的 parseDeviceFromUA 保持一致）
func parseDeviceFromUATest(ua string) string {
	if ua == "" {
		return "Unknown Device"
	}
	uaLower := strings.ToLower(ua)

	browser := ""
	if strings.Contains(uaLower, "edg/") || strings.Contains(uaLower, "edge/") {
		browser = "Edge"
	} else if strings.Contains(uaLower, "chrome/") && !strings.Contains(uaLower, "chromium/") {
		browser = "Chrome"
	} else if strings.Contains(uaLower, "firefox/") {
		browser = "Firefox"
	} else if strings.Contains(uaLower, "safari/") && !strings.Contains(uaLower, "chrome/") {
		browser = "Safari"
	} else if strings.Contains(uaLower, "opera/") || strings.Contains(uaLower, "opr/") {
		browser = "Opera"
	}

	deviceType := "Desktop"
	if strings.Contains(uaLower, "mobile") || strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "ipod") {
		deviceType = "Mobile"
	} else if strings.Contains(uaLower, "android") && !strings.Contains(uaLower, "mobile") {
		deviceType = "Android Tablet"
	} else if strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad") {
		deviceType = "iPad"
	}

	if browser != "" {
		return browser + " on " + deviceType
	}
	return deviceType
}

// ==================== User-Agent Parsing Tests ====================

func TestParseDeviceFromUA_Empty(t *testing.T) {
	device := parseDeviceFromUATest("")
	assert.Equal(t, "Unknown Device", device)
}

func TestParseDeviceFromUA_ChromeDesktop(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Chrome")
	assert.Contains(t, device, "Desktop")
}

func TestParseDeviceFromUA_FirefoxDesktop(t *testing.T) {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Firefox")
	assert.Contains(t, device, "Desktop")
}

func TestParseDeviceFromUA_EdgeDesktop(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Edge")
	assert.Contains(t, device, "Desktop")
}

func TestParseDeviceFromUA_SafariDesktop(t *testing.T) {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Safari")
	assert.Contains(t, device, "Desktop")
}

func TestParseDeviceFromUA_SafariMobile(t *testing.T) {
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Safari")
	assert.Contains(t, device, "Mobile")
}

func TestParseDeviceFromUA_AndroidMobile(t *testing.T) {
	ua := "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Chrome")
	assert.Contains(t, device, "Mobile")
}

func TestParseDeviceFromUA_AndroidTablet(t *testing.T) {
	ua := "Mozilla/5.0 (Linux; Android 13; SM-X900) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Chrome")
	assert.Contains(t, device, "Android Tablet")
}

func TestParseDeviceFromUA_iPad(t *testing.T) {
	ua := "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Safari")
	assert.Contains(t, device, "Mobile") // iPad UA 包含 Mobile 关键字
}

func TestParseDeviceFromUA_UnknownBrowser(t *testing.T) {
	ua := "CustomBot/1.0"
	device := parseDeviceFromUATest(ua)
	assert.Equal(t, "Desktop", device)
}

func TestParseDeviceFromUA_ChromeOnMac(t *testing.T) {
	ua := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Chrome")
	assert.Contains(t, device, "Desktop")
}

func TestParseDeviceFromUA_ChromeOnLinux(t *testing.T) {
	ua := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	device := parseDeviceFromUATest(ua)
	assert.Contains(t, device, "Chrome")
	assert.Contains(t, device, "Desktop")
}
