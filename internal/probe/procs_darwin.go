//go:build darwin

package probe

import (
	"bytes"

	"golang.org/x/sys/unix"
)

// processNames 通过 sysctl kern.proc.all 读取内核进程表（App Sandbox 内同样可用，无需执行 ps）
func processNames() (map[string]bool, bool) {
	procs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, false
	}
	names := make(map[string]bool, len(procs))
	for i := range procs {
		names[unix.ByteSliceToString(procs[i].Proc.P_comm[:])] = true
	}
	return names, true
}

// runningAppBundles 返回当前用户正在运行的进程所属的最外层 .app 包路径（通过 kern.procargs2 读取可执行文件路径）。
// 一个进程都读不到时（如沙盒限制）返回 ok=false，调用方据此保守处理。
func runningAppBundles() (map[string]bool, bool) {
	procs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, false
	}
	uid := uint32(unix.Getuid())
	bundles := make(map[string]bool)
	readable := false
	for i := range procs {
		if procs[i].Eproc.Ucred.Uid != uid {
			continue
		}
		buf, err := unix.SysctlRaw("kern.procargs2", int(procs[i].Proc.P_pid))
		if err != nil || len(buf) <= 4 {
			continue
		}
		readable = true
		exe := buf[4:]
		if n := bytes.IndexByte(exe, 0); n >= 0 {
			exe = exe[:n]
		}
		if b := outerAppBundle(string(exe)); b != "" {
			bundles[b] = true
		}
	}
	return bundles, readable
}

// processArgs 返回当前用户进程中可执行文件路径满足 match 的进程的完整参数（argv）。
// 一个进程都读不到时返回 ok=false，调用方据此保守处理
func processArgs(match func(exe string) bool) ([][]string, bool) {
	procs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, false
	}
	uid := uint32(unix.Getuid())
	var out [][]string
	readable := false
	for i := range procs {
		if procs[i].Eproc.Ucred.Uid != uid {
			continue
		}
		buf, err := unix.SysctlRaw("kern.procargs2", int(procs[i].Proc.P_pid))
		if err != nil {
			continue
		}
		readable = true
		if exe, args := parseProcArgs2(buf); exe != "" && match(exe) {
			out = append(out, args)
		}
	}
	return out, readable
}
