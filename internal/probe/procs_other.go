//go:build !darwin

package probe

// processNames 非 macOS 平台暂不读取进程表，调用方据此保守视为运行中
func processNames() (map[string]bool, bool) {
	return nil, false
}
