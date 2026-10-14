package probe

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
	"devlemon/internal/rules"
)

// RuleProbe 按声明式规则 (internal/rules) 扫描一个分组：每条规则的目录聚合为一个清理项
type RuleProbe struct {
	group    string
	groupID  string
	category model.Category
	titleEn  string
	titleZh  string
}

// NewCacheProbe 开发语言与包管理器缓存（规则分组 package_caches）
func NewCacheProbe() *RuleProbe {
	return &RuleProbe{
		group: rules.GroupPackageCaches, groupID: "package_caches", category: model.CategoryPackageCache,
		titleEn: "📦 Package Manager & Developer Caches", titleZh: "📦 开发语言与包管理器缓存",
	}
}

// NewAppCacheProbe 应用垃圾与缓存（规则分组 app_caches）
func NewAppCacheProbe() *RuleProbe {
	return &RuleProbe{
		group: rules.GroupAppCaches, groupID: "app_caches", category: model.CategoryAppCache,
		titleEn: "📱 Application Caches & Garbage", titleZh: "📱 应用垃圾与缓存（Xcode、VSCode、微信、办公影音）",
	}
}

func (p *RuleProbe) Name() string {
	return "Rule Probe (" + p.group + ")"
}

func (p *RuleProbe) Category() model.Category {
	return p.category
}

func (p *RuleProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       p.groupID,
		Title:    i18n.T(p.titleEn, p.titleZh),
		Category: p.category,
		Items:    make([]*model.Item, 0),
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return group, nil
	}
	set := loadedRules()
	if set == nil {
		return group, nil
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, r := range set.ForGroup(p.group) {
		r := r
		if !r.AppliesTo(runtime.GOOS) {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			item := ruleItem(r, home, p.category)
			if item == nil {
				return
			}
			mu.Lock()
			group.Items = append(group.Items, item)
			group.TotalSizeBytes += item.SizeBytes
			group.TotalReclaimableBytes += item.SizeBytes
			mu.Unlock()
		}()
	}
	wg.Wait()
	return group, nil
}

// ruleItem 统计规则下所有已存在目录的体积；有 clean_command 时由命令负责清理（如 go clean -cache）
func ruleItem(r rules.Rule, home string, category model.Category) *model.Item {
	item := newCacheItem(r.ID, r.Title.String(), r.Description.String(), r.ExpandPaths(home), r.MinBytes(), model.RiskLevel(r.Risk))
	if item == nil {
		return nil
	}
	item.Category = category
	if len(r.CleanCommand) > 0 {
		item.CleanType = model.CleanTypeCommand
		item.CleanCommand = r.CleanCommand
		item.CleanPaths = nil
		item.CleanPath = item.Path
	}
	return item
}

var (
	rulesOnce sync.Once
	rulesSet  *rules.Set
)

// loadedRules 每次扫描加载一次规则（由 ResetScanCaches 重置）；内置规则错误属于程序缺陷，打印后跳过规则扫描
func loadedRules() *rules.Set {
	rulesOnce.Do(func() {
		set, err := rules.Load()
		if err != nil {
			os.Stderr.WriteString("devlemon: failed to load built-in rules: " + err.Error() + "\n")
			return
		}
		for _, w := range set.Warnings {
			os.Stderr.WriteString("devlemon: rule warning: " + w + "\n")
		}
		rulesSet = set
	})
	return rulesSet
}

// ruleCoverage 规则覆盖的目录，供其他探针去重：
//   - cachesNames：~/Library/Caches 下被任一规则覆盖（本身或其子目录）的一级目录名
//   - bundleIDs：规则路径中出现的 Bundle ID（小写），如容器 com.tencent.xinWeChat
func ruleCoverage(home string) (cachesNames, bundleIDs map[string]bool) {
	cachesNames, bundleIDs = map[string]bool{}, map[string]bool{}
	set := loadedRules()
	if set == nil {
		return
	}
	caches := filepath.Join(home, "Library", "Caches") + string(filepath.Separator)
	lib := filepath.Join(home, "Library") + string(filepath.Separator)
	for _, r := range set.Rules {
		for _, p := range r.ExpandPaths(home) {
			if strings.HasPrefix(p, caches) {
				cachesNames[strings.SplitN(strings.TrimPrefix(p, caches), string(filepath.Separator), 2)[0]] = true
			}
			if strings.HasPrefix(p, lib) {
				for _, seg := range strings.Split(strings.TrimPrefix(p, lib), string(filepath.Separator)) {
					if looksLikeBundleID(seg) {
						bundleIDs[strings.ToLower(seg)] = true
					}
				}
			}
		}
	}
	return
}
