//go:build darwin

package probe

import "golang.org/x/sys/unix"

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
