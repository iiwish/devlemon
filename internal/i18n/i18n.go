package i18n

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var isChinese = detectChinese()

// IsChinese returns whether Chinese is the active language
func IsChinese() bool {
	return isChinese
}

// SetChinese allows programmatic override
func SetChinese(v bool) {
	isChinese = v
}

// T returns zh when the current environment is Chinese, otherwise returns en
func T(en, zh string) string {
	if isChinese {
		return zh
	}
	return en
}

func detectChinese() bool {
	// 1. Explicit user override
	if lang := os.Getenv("DEVLEMON_LANG"); lang != "" {
		return isZh(lang)
	}
	if lang := os.Getenv("DEVLEMON_LOCALE"); lang != "" {
		return isZh(lang)
	}

	// 2. Darwin system settings check
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("defaults", "read", "-g", "AppleLocale").Output(); err == nil {
			locale := strings.TrimSpace(string(out))
			if isZh(locale) {
				return true
			}
			if locale != "" && !isZh(locale) {
				return false
			}
		}
		if out, err := exec.Command("defaults", "read", "-g", "AppleLanguages").Output(); err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "\"") {
					return isZh(line)
				}
			}
		}
	}

	// 3. POSIX environment variables
	for _, envKey := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if val := os.Getenv(envKey); val != "" {
			if isZh(val) {
				return true
			}
			if !strings.HasPrefix(val, "C") && val != "POSIX" {
				return false
			}
		}
	}

	// Default fallback: English
	return false
}

func isZh(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "zh") ||
		strings.Contains(lower, "hans") ||
		strings.Contains(lower, "hant")
}
