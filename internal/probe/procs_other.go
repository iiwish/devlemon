//go:build !darwin

package probe

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// processNames 非 macOS 平台暂不读取进程表，调用方据此保守视为运行中
func processNames() (map[string]bool, bool) {
	return nil, false
}

// runningAppBundles 非 macOS 平台没有 .app 包，调用方据此保守视为运行中
func runningAppBundles() (map[string]bool, bool) {
	return nil, false
}

// processArgs 通过 /proc/<pid>/cmdline 读取进程参数（Linux）；没有 /proc 时返回 ok=false
func processArgs(match func(exe string) bool) ([][]string, bool) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, false
	}
	var out [][]string
	readable := false
	for _, e := range entries {
		if strings.Trim(e.Name(), "0123456789") != "" {
			continue
		}
		data, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline"))
		if err != nil || len(data) == 0 {
			continue
		}
		readable = true
		args := strings.Split(string(bytes.TrimRight(data, "\x00")), "\x00")
		exe, err := os.Readlink(filepath.Join("/proc", e.Name(), "exe"))
		if err != nil {
			exe = args[0]
		}
		if match(exe) {
			out = append(out, args)
		}
	}
	return out, readable
}
