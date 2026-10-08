package i18n

import (
	"os"
	"testing"
)

func TestI18n(t *testing.T) {
	// Test override
	SetChinese(true)
	if T("Hello", "你好") != "你好" {
		t.Errorf("Expected '你好', got '%s'", T("Hello", "你好"))
	}

	SetChinese(false)
	if T("Hello", "你好") != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", T("Hello", "你好"))
	}

	// Restore detection
	isChinese = detectChinese()
}

func TestDetectWithEnv(t *testing.T) {
	os.Setenv("DEVLEMON_LANG", "en")
	if detectChinese() != false {
		t.Errorf("Expected false for DEVLEMON_LANG=en")
	}

	os.Setenv("DEVLEMON_LANG", "zh_CN")
	if detectChinese() != true {
		t.Errorf("Expected true for DEVLEMON_LANG=zh_CN")
	}
	os.Unsetenv("DEVLEMON_LANG")
}
