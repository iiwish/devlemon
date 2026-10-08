package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

type SimulatorProbe struct{}

func NewSimulatorProbe() *SimulatorProbe {
	return &SimulatorProbe{}
}

func (p *SimulatorProbe) Name() string {
	return "iOS Simulator Active Shield Probe"
}

func (p *SimulatorProbe) Category() model.Category {
	return model.CategorySimulator
}

type simctlDevice struct {
	UDID  string `json:"udid"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type simctlListOutput struct {
	Devices map[string][]simctlDevice `json:"devices"`
}

func (p *SimulatorProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       "ios_simulators",
		Title:    i18n.T("🛠️ iOS Simulator Environment (Active Shield Protected)", "🛠️ iOS 模拟器环境（带运行态避让保护）"),
		Category: model.CategorySimulator,
		Items:    make([]*model.Item, 0),
	}

	// 检查 xcrun 是否可用
	if _, err := exec.LookPath("xcrun"); err != nil {
		return group, nil
	}

	cmd := exec.CommandContext(ctx, "xcrun", "simctl", "list", "devices", "-j")
	out, err := cmd.Output()
	if err != nil {
		return group, nil
	}

	var simData simctlListOutput
	if err := json.Unmarshal(out, &simData); err != nil {
		return group, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return group, nil
	}
	devicesDir := filepath.Join(home, "Library", "Developer", "CoreSimulator", "Devices")

	for _, devList := range simData.Devices {
		for _, dev := range devList {
			devPath := filepath.Join(devicesDir, dev.UDID)
			info, err := os.Stat(devPath)
			if err != nil || !info.IsDir() {
				continue
			}

			size := FastDirSize(devPath)

			// 1. 如果设备正处于运行状态 (Booted) -> 强制列入保护名单！
			if dev.State == "Booted" {
				group.Items = append(group.Items, &model.Item{
					ID:            "sim_" + dev.UDID,
					Title:         fmt.Sprintf(i18n.T("Simulator: %s", "模拟器: %s"), dev.Name),
					Description:   i18n.T("Running in foreground/background, active safety shield applied", "当前测试设备正在前台/后台运行中，已开启安全避让保护"),
					Path:          devPath,
					SizeBytes:     size,
					SizeFormatted: model.FormatBytes(size),
					Risk:          model.RiskSafe,
					Category:      model.CategorySimulator,
					IsProtected:   true,
					ProtectReason: i18n.T("🟢 Running (Booted), locked by safety shield to prevent interruption", "🟢 正在运行中 (Booted)，已自动加锁保护，绝不中断测试"),
				})
				group.TotalSizeBytes += size
				continue
			}

			// 2. 如果是已关机设备且占用超过 100MB（证明有历史沙盒或测试应用残留）
			if size > 100*1024*1024 {
				group.Items = append(group.Items, &model.Item{
					ID:            "sim_" + dev.UDID,
					Title:         fmt.Sprintf(i18n.T("Shutdown Device: %s", "已关机设备: %s"), dev.Name),
					Description:   i18n.T("Historical test app sandbox data and cache, safe to erase", "历史测试残留数据与应用沙盒缓存，重置后恢复初始洁净状态"),
					Path:          devPath,
					SizeBytes:     size,
					SizeFormatted: model.FormatBytes(size),
					Risk:          model.RiskCaution,
					Category:      model.CategorySimulator,
					IsProtected:   false,
					CleanType:     model.CleanTypeCommand,
					CleanCommand:  []string{"xcrun", "simctl", "erase", dev.UDID},
				})
				group.TotalSizeBytes += size
				group.TotalReclaimableBytes += size
			}
		}
	}

	return group, nil
}
