// Package rules 声明式清理规则：一条规则把一个或多个目录描述为一个清理项。
// 内置规则以 YAML 打包在 builtin/ 中；用户可在 ~/.config/devlemon/rules.d/*.yaml 中新增、覆盖或禁用规则。
// 只适合「目录即缓存」的简单场景；需要运行态检测、版本比较等判断的规则仍用 Go 实现。
package rules

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"devlemon/internal/i18n"
)

//go:embed builtin/*.yaml
var builtinFS embed.FS

// 规则所属分组，决定扫描结果中归入哪一组
const (
	GroupPackageCaches = "package_caches"
	GroupAppCaches     = "app_caches"
)

var knownGroups = map[string]bool{GroupPackageCaches: true, GroupAppCaches: true}

var knownRisks = map[string]bool{"safe": true, "rebuildable": true, "caution": true}

var knownOS = map[string]bool{"darwin": true, "linux": true, "windows": true}

// 路径前缀：~/ 为用户主目录；{darwin_user_cache}/ 为 macOS 每用户缓存目录（$TMPDIR 同级的 C/）
const (
	homePrefix      = "~/"
	userCachePrefix = "{darwin_user_cache}/"
)

// Text 中英文文本，缺少中文时回退为英文
type Text struct {
	En string `yaml:"en"`
	Zh string `yaml:"zh"`
}

func (t Text) String() string {
	if t.Zh == "" {
		return t.En
	}
	return i18n.T(t.En, t.Zh)
}

// Rule 一条清理规则
type Rule struct {
	ID           string   `yaml:"id"`
	Title        Text     `yaml:"title"`
	Description  Text     `yaml:"description"`
	Paths        []string `yaml:"paths"`
	Risk         string   `yaml:"risk"`
	MinSize      string   `yaml:"min_size"`
	CleanCommand []string `yaml:"clean_command"`
	OS           []string `yaml:"os"`
	Enabled      *bool    `yaml:"enabled"`

	Group  string `yaml:"-"` // 来自文件的 group 字段
	Source string `yaml:"-"` // 规则来源文件，便于排查
}

type file struct {
	Group string `yaml:"group"`
	Rules []Rule `yaml:"rules"`
}

var idRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_]*$`)

// protectedHomePaths 绝不允许作为清理路径的主目录下目录（只禁止精确匹配，其子目录可以）
var protectedHomePaths = map[string]bool{
	"Library": true, "Documents": true, "Desktop": true, "Downloads": true, "Pictures": true,
	"Movies": true, "Music": true, "Applications": true, "Public": true, "Developer": true,
	".ssh": true, ".gnupg": true, ".config": true, ".local": true, ".local/share": true,
	".cache": true, ".aws": true, ".kube": true, ".docker": true, ".Trash": true,
	"Library/Application Support": true, "Library/Caches": true, "Library/Containers": true,
	"Library/Group Containers": true, "Library/Preferences": true, "Library/Developer": true,
	"Library/Mobile Documents": true, "Library/Keychains": true, "Library/Mail": true,
}

// Enabled 规则是否启用（未写 enabled 视为启用）
func (r Rule) IsEnabled() bool {
	return r.Enabled == nil || *r.Enabled
}

// Validate 校验规则合法性。所有规则（含内置）都必须通过，路径限制是删除前的第一道防线
func (r Rule) Validate() error {
	if !idRe.MatchString(r.ID) {
		return fmt.Errorf("invalid id %q (lowercase letters, digits and _ only)", r.ID)
	}
	if !knownGroups[r.Group] {
		return fmt.Errorf("%s: the file must set group to %s or %s", r.ID, GroupPackageCaches, GroupAppCaches)
	}
	if r.Title.En == "" {
		return fmt.Errorf("%s: title.en is required", r.ID)
	}
	if !knownRisks[r.Risk] {
		return fmt.Errorf("%s: risk must be safe, rebuildable or caution (got %q)", r.ID, r.Risk)
	}
	if _, err := ParseSize(r.MinSize); err != nil {
		return fmt.Errorf("%s: %w", r.ID, err)
	}
	for _, o := range r.OS {
		if !knownOS[o] {
			return fmt.Errorf("%s: unknown os %q", r.ID, o)
		}
	}
	if len(r.Paths) == 0 {
		return fmt.Errorf("%s: at least one path is required", r.ID)
	}
	for _, p := range r.Paths {
		if err := validatePath(p); err != nil {
			return fmt.Errorf("%s: %w", r.ID, err)
		}
	}
	if len(r.CleanCommand) > 0 && strings.TrimSpace(r.CleanCommand[0]) == "" {
		return fmt.Errorf("%s: clean_command must start with a program name", r.ID)
	}
	return nil
}

func validatePath(p string) error {
	var rest string
	switch {
	case strings.HasPrefix(p, homePrefix):
		rest = strings.TrimPrefix(p, homePrefix)
	case strings.HasPrefix(p, userCachePrefix):
		rest = strings.TrimPrefix(p, userCachePrefix)
	default:
		return fmt.Errorf("path %q must start with ~/ or %s", p, userCachePrefix)
	}
	if strings.ContainsAny(rest, "*?") {
		return fmt.Errorf("path %q: wildcards are not supported", p)
	}
	if strings.Contains(rest, "\\") {
		return fmt.Errorf("path %q: use / as separator", p)
	}
	clean := filepath.ToSlash(filepath.Clean(rest))
	if clean == "." || clean == "" || clean != strings.TrimSuffix(rest, "/") {
		return fmt.Errorf("path %q must be a normalized path below its prefix (no empty, . or .. segments)", p)
	}
	for _, seg := range strings.Split(clean, "/") {
		if seg == ".." || seg == "." {
			return fmt.Errorf("path %q must not contain . or .. segments", p)
		}
	}
	if strings.HasPrefix(p, homePrefix) && protectedHomePaths[clean] {
		return fmt.Errorf("path %q is a protected folder and cannot be cleaned as a whole", p)
	}
	return nil
}

// MinBytes 规则的最小体积（未设置时默认 10MB）
func (r Rule) MinBytes() int64 {
	n, _ := ParseSize(r.MinSize)
	return n
}

var sizeRe = regexp.MustCompile(`^\s*(\d+(?:\.\d+)?)\s*([KMG]?B)?\s*$`)

// ParseSize 解析 "10MB"、"512KB"、"1.5GB"、"0" 等体积写法，空字符串表示默认 10MB
func ParseSize(s string) (int64, error) {
	if strings.TrimSpace(s) == "" {
		return 10 * 1024 * 1024, nil
	}
	m := sizeRe.FindStringSubmatch(strings.ToUpper(s))
	if m == nil {
		return 0, fmt.Errorf("invalid min_size %q (examples: 10MB, 512KB, 1GB)", s)
	}
	v, _ := strconv.ParseFloat(m[1], 64)
	mult := map[string]float64{"": 1, "B": 1, "KB": 1024, "MB": 1024 * 1024, "GB": 1024 * 1024 * 1024}[m[2]]
	return int64(v * mult), nil
}

// AppliesTo 规则是否适用于当前操作系统
func (r Rule) AppliesTo(goos string) bool {
	if len(r.OS) == 0 {
		return true
	}
	for _, o := range r.OS {
		if o == goos {
			return true
		}
	}
	return false
}

// ExpandPaths 把规则路径展开为绝对路径；某个前缀变量在本机不可用时跳过该路径
func (r Rule) ExpandPaths(home string) []string {
	var out []string
	for _, p := range r.Paths {
		switch {
		case strings.HasPrefix(p, homePrefix):
			out = append(out, filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(p, homePrefix))))
		case strings.HasPrefix(p, userCachePrefix):
			if dir := DarwinUserCacheDir(); dir != "" {
				out = append(out, filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(p, userCachePrefix))))
			}
		}
	}
	return out
}

// DarwinUserCacheDir macOS 每用户缓存目录（/var/folders/xx/yyy/C）；非 macOS 或 TMPDIR 非标准时返回空
func DarwinUserCacheDir() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	tmp := filepath.Clean(os.TempDir())
	if !strings.HasPrefix(tmp, "/var/folders/") && !strings.HasPrefix(tmp, "/private/var/folders/") {
		return ""
	}
	return filepath.Join(filepath.Dir(tmp), "C")
}

// ParseFile 解析一个规则文件。未知字段直接报错，避免拼写错误被静默忽略
func ParseFile(data []byte, source string) ([]Rule, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var f file
	if err := dec.Decode(&f); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", source, err)
	}
	// 只用于禁用规则的文件可以省略 group；启用的规则缺少 group 会在 Validate 中报错
	if f.Group != "" && !knownGroups[f.Group] {
		return nil, fmt.Errorf("%s: group must be %s or %s (got %q)", source, GroupPackageCaches, GroupAppCaches, f.Group)
	}
	seen := map[string]bool{}
	for i := range f.Rules {
		f.Rules[i].Group = f.Group
		f.Rules[i].Source = source
		if seen[f.Rules[i].ID] {
			return nil, fmt.Errorf("%s: duplicate rule id %q", source, f.Rules[i].ID)
		}
		seen[f.Rules[i].ID] = true
	}
	return f.Rules, nil
}

// Set 一次加载得到的全部规则
type Set struct {
	Rules    []Rule   // 已启用且校验通过的规则
	Warnings []string // 用户规则中的问题（无效规则被跳过，不影响其他规则）
}

// ForGroup 返回某分组下的规则
func (s *Set) ForGroup(group string) []Rule {
	var out []Rule
	for _, r := range s.Rules {
		if r.Group == group {
			out = append(out, r)
		}
	}
	return out
}

// Builtin 解析内置规则；内置规则必须全部合法，否则视为程序错误
func Builtin() ([]Rule, error) {
	entries, err := builtinFS.ReadDir("builtin")
	if err != nil {
		return nil, err
	}
	var all []Rule
	seen := map[string]string{}
	for _, e := range entries {
		name := "builtin/" + e.Name()
		data, err := builtinFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		rs, err := ParseFile(data, name)
		if err != nil {
			return nil, err
		}
		for _, r := range rs {
			if err := r.Validate(); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			if prev, ok := seen[r.ID]; ok {
				return nil, fmt.Errorf("rule id %q defined in both %s and %s", r.ID, prev, name)
			}
			seen[r.ID] = name
			all = append(all, r)
		}
	}
	return all, nil
}

// UserRulesDir 用户规则目录：$DEVLEMON_RULES_DIR，否则 $XDG_CONFIG_HOME/devlemon/rules.d，否则 ~/.config/devlemon/rules.d
func UserRulesDir() string {
	if d := os.Getenv("DEVLEMON_RULES_DIR"); d != "" {
		return d
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "devlemon", "rules.d")
}

// Load 加载内置规则与用户规则。用户规则按文件名顺序应用：同 id 覆盖内置规则，enabled: false 禁用规则。
// 用户规则有问题时跳过该规则或文件并记入 Warnings，不影响扫描
func Load() (*Set, error) {
	builtin, err := Builtin()
	if err != nil {
		return nil, err
	}
	byID := map[string]Rule{}
	var order []string
	for _, r := range builtin {
		byID[r.ID] = r
		order = append(order, r.ID)
	}

	set := &Set{}
	dir := UserRulesDir()
	files, _ := filepath.Glob(filepath.Join(dir, "*.yaml"))
	ymls, _ := filepath.Glob(filepath.Join(dir, "*.yml"))
	files = append(files, ymls...)
	sort.Strings(files)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			set.Warnings = append(set.Warnings, err.Error())
			continue
		}
		rs, err := ParseFile(data, path)
		if err != nil {
			set.Warnings = append(set.Warnings, err.Error())
			continue
		}
		for _, r := range rs {
			if !r.IsEnabled() {
				// 仅凭 id 即可禁用规则，无需重复填写其他字段
				if !idRe.MatchString(r.ID) {
					set.Warnings = append(set.Warnings, fmt.Sprintf("%s: invalid id %q", path, r.ID))
					continue
				}
				if _, ok := byID[r.ID]; !ok {
					set.Warnings = append(set.Warnings, fmt.Sprintf("%s: cannot disable unknown rule %q", path, r.ID))
				}
				delete(byID, r.ID)
				continue
			}
			if err := r.Validate(); err != nil {
				set.Warnings = append(set.Warnings, fmt.Sprintf("%s: %v", path, err))
				continue
			}
			if _, ok := byID[r.ID]; !ok {
				order = append(order, r.ID)
			}
			byID[r.ID] = r
		}
	}
	for _, id := range order {
		if r, ok := byID[id]; ok {
			set.Rules = append(set.Rules, r)
			delete(byID, id) // 同一 id 只输出一次
		}
	}
	return set, nil
}
