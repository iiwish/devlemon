package probe

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

type CacheProbe struct{}

func NewCacheProbe() *CacheProbe {
	return &CacheProbe{}
}

func (p *CacheProbe) Name() string {
	return "Developer & System Cache Probe"
}

func (p *CacheProbe) Category() model.Category {
	return model.CategoryPackageCache
}

type cacheTarget struct {
	id          string
	title       string
	description string
	subpath     string
	absPath     string   // 不在主目录下的绝对路径（优先于 subpath）
	subpaths    []string // 多个目录聚合为一项（优先于 subpath / absPath）
	risk        model.RiskLevel
	cleanCmd    []string
}

func (p *CacheProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	targets := []cacheTarget{
		{
			id:          "go_build_cache",
			title:       i18n.T("Go Build Cache", "Go 编译中间缓存"),
			description: i18n.T("Temporary build object files produced by Go compiler", "Go 编译器生成的临时目标文件，删除后随时按需自动重新生成"),
			subpath:     filepath.Join("Library", "Caches", "go-build"),
			risk:        model.RiskRebuildable,
			cleanCmd:    []string{"go", "clean", "-cache"},
		},
		{
			id:          "go_mod_cache",
			title:       i18n.T("Go Module Download Cache", "Go 模块下载缓存"),
			description: i18n.T("Locally downloaded third-party Go dependencies (pkg/mod)", "本地下载缓存的第三方 Go 依赖包 (pkg/mod)"),
			subpath:     filepath.Join("go", "pkg", "mod"),
			risk:        model.RiskRebuildable,
			cleanCmd:    []string{"go", "clean", "-modcache"},
		},
		{
			id:          "npm_cache",
			title:       i18n.T("npm Package Cache", "npm 离线安装包缓存"),
			description: i18n.T("Global npm tarball cache (~/.npm), safe to purge", "npm 全局安装包缓存 (~/.npm)，按需重新拉取"),
			subpath:     ".npm",
			risk:        model.RiskRebuildable,
			cleanCmd:    []string{"npm", "cache", "clean", "--force"},
		},
		{
			id:          "pnpm_cache",
			title:       i18n.T("pnpm Store Cache", "pnpm 临时存储缓存"),
			description: i18n.T("Local client download cache for pnpm", "pnpm 客户端本地下载缓存"),
			subpath:     filepath.Join("Library", "Caches", "pnpm"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "uv_cache",
			title:       i18n.T("Python uv Package Cache", "Python uv 高速包缓存"),
			description: i18n.T("Download and prebuilt wheel cache for uv toolchain", "Python uv 工具链下载与预构建 wheel 缓存"),
			subpath:     filepath.Join(".cache", "uv"),
			risk:        model.RiskRebuildable,
			cleanCmd:    []string{"uv", "cache", "clean"},
		},
		{
			id:          "pip_cache",
			title:       i18n.T("Python pip Download Cache", "Python pip 下载缓存"),
			description: i18n.T("Offline download cache for pip packages", "pip 安装包离线缓存"),
			subpath:     filepath.Join("Library", "Caches", "pip"),
			risk:        model.RiskRebuildable,
			cleanCmd:    []string{"pip", "cache", "purge"},
		},
		{
			id:          "homebrew_cache",
			title:       i18n.T("Homebrew Bottle & Download Cache", "Homebrew 安装包与 Bottle 缓存"),
			description: i18n.T("Cached bottle downloads and outdated package archives", "Brew 下载的历史版本二进制包与安装归档"),
			subpath:     filepath.Join("Library", "Caches", "Homebrew"),
			risk:        model.RiskRebuildable,
			cleanCmd:    []string{"brew", "cleanup", "-s"},
		},
		{
			id:          "playwright_browsers",
			title:       i18n.T("Playwright Headless Browsers", "Playwright 自动化无头浏览器包"),
			description: i18n.T("Chromium/WebKit standalone browser binaries for E2E tests", "用于 E2E 自动化测试的 Chromium/WebKit 独立浏览器包"),
			subpath:     filepath.Join("Library", "Caches", "ms-playwright"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "puppeteer_browsers",
			title:       i18n.T("Puppeteer Headless Browsers", "Puppeteer 无头测试浏览器缓存"),
			description: i18n.T("Downloaded Chromium browser binaries for Puppeteer", "Puppeteer 离线测试浏览器运行时 (~/.cache/puppeteer)"),
			subpath:     filepath.Join(".cache", "puppeteer"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "cocoapods_cache",
			title:       i18n.T("CocoaPods Dependency Cache", "CocoaPods 依赖下载缓存"),
			description: i18n.T("Downloaded pods archives and specs (~/Library/Caches/CocoaPods)", "CocoaPods 本地下载归档包，删除无副作用"),
			subpath:     filepath.Join("Library", "Caches", "CocoaPods"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "ios_firmware",
			title:       i18n.T("iOS / iPadOS Firmware Downloads", "iOS / iPadOS 固件安装包 (.ipsw)"),
			description: i18n.T("Firmware downloaded by Finder when updating or restoring devices; re-downloaded when needed", "Finder 更新或恢复设备时下载的系统固件，用完即可删除，需要时重新下载"),
			subpath:     filepath.Join("Library", "iTunes", "iPhone Software Updates"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "ipados_firmware",
			title:       i18n.T("iPad Firmware Downloads", "iPad 固件安装包 (.ipsw)"),
			description: i18n.T("Firmware downloaded by Finder when updating or restoring devices; re-downloaded when needed", "Finder 更新或恢复设备时下载的系统固件，用完即可删除，需要时重新下载"),
			subpath:     filepath.Join("Library", "iTunes", "iPad Software Updates"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "cargo_cache",
			title:       i18n.T("Cargo Crate Download Cache", "Cargo 依赖下载缓存"),
			description: i18n.T("Downloaded crate archives, extracted sources and git dependencies in ~/.cargo; Cargo downloads them again on the next build", "~/.cargo 中下载的 crate 压缩包、解压源码与 git 依赖，下次构建时 Cargo 会重新下载"),
			subpaths: []string{
				filepath.Join(".cargo", "registry", "cache"),
				filepath.Join(".cargo", "registry", "src"),
				filepath.Join(".cargo", "git", "checkouts"),
				filepath.Join(".cargo", "git", "db"),
			},
			risk: model.RiskRebuildable,
		},
		{
			id:          "gradle_caches",
			title:       i18n.T("Gradle Dependency & Build Cache", "Gradle 依赖与构建缓存"),
			description: i18n.T("Downloaded dependencies and build cache in ~/.gradle/caches; rebuilt on the next Gradle build (stop running Gradle daemons first)", "~/.gradle/caches 中的依赖与构建缓存，下次构建时自动重新下载（清理前请先停止正在运行的 Gradle 守护进程）"),
			subpath:     filepath.Join(".gradle", "caches"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "maven_repository",
			title:       i18n.T("Maven Local Repository", "Maven 本地依赖仓库"),
			description: i18n.T("Dependencies downloaded by Maven (~/.m2/repository); downloaded again on the next build. Locally installed (mvn install) artifacts are also removed", "Maven 下载的依赖 (~/.m2/repository)，下次构建时重新下载；通过 mvn install 安装到本地的构件也会一并删除"),
			subpath:     filepath.Join(".m2", "repository"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "yarn_cache",
			title:       i18n.T("Yarn Package Cache", "Yarn 依赖缓存"),
			description: i18n.T("Global package cache of Yarn; packages are downloaded again on the next install", "Yarn 全局依赖缓存，下次安装时重新下载"),
			subpaths: []string{
				filepath.Join("Library", "Caches", "Yarn"),
				filepath.Join(".cache", "yarn"),
				filepath.Join(".yarn", "berry", "cache"),
			},
			risk: model.RiskRebuildable,
		},
		{
			id:          "bun_cache",
			title:       i18n.T("Bun Package Cache", "Bun 依赖缓存"),
			description: i18n.T("Global install cache of Bun (~/.bun/install/cache)", "Bun 全局安装缓存 (~/.bun/install/cache)，下次安装时重新下载"),
			subpath:     filepath.Join(".bun", "install", "cache"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "deno_cache",
			title:       i18n.T("Deno Module Cache", "Deno 模块缓存"),
			description: i18n.T("Remote modules and compiled code cached by Deno", "Deno 缓存的远程模块与编译产物，下次运行时重新下载"),
			subpaths: []string{
				filepath.Join("Library", "Caches", "deno"),
				filepath.Join(".cache", "deno"),
			},
			risk: model.RiskRebuildable,
		},
		{
			id:          "composer_cache",
			title:       i18n.T("Composer Package Cache", "Composer 依赖缓存"),
			description: i18n.T("Package download cache of PHP Composer", "PHP Composer 的依赖下载缓存"),
			subpaths: []string{
				filepath.Join("Library", "Caches", "composer"),
				filepath.Join(".composer", "cache"),
				filepath.Join(".cache", "composer"),
			},
			risk: model.RiskRebuildable,
		},
		{
			id:          "huggingface_cache",
			title:       i18n.T("Hugging Face Model & Dataset Cache", "Hugging Face 模型与数据集缓存"),
			description: i18n.T("Models and datasets downloaded by transformers / diffusers / datasets (~/.cache/huggingface/hub). Downloaded again on next use, which can be slow for large models", "transformers / diffusers / datasets 下载的模型与数据集 (~/.cache/huggingface/hub)，下次使用时重新下载，大模型可能较慢"),
			subpaths: []string{
				filepath.Join(".cache", "huggingface", "hub"),
				filepath.Join(".cache", "huggingface", "datasets"),
			},
			risk: model.RiskRebuildable,
		},
		{
			id:          "torch_hub_cache",
			title:       i18n.T("PyTorch Hub Model Cache", "PyTorch Hub 模型缓存"),
			description: i18n.T("Pretrained weights downloaded via torch.hub (~/.cache/torch); downloaded again on next use", "通过 torch.hub 下载的预训练权重 (~/.cache/torch)，下次使用时重新下载"),
			subpath:     filepath.Join(".cache", "torch"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "whisper_cache",
			title:       i18n.T("Whisper Model Cache", "Whisper 语音模型缓存"),
			description: i18n.T("Speech recognition models downloaded by openai-whisper (~/.cache/whisper)", "openai-whisper 下载的语音识别模型 (~/.cache/whisper)，下次使用时重新下载"),
			subpath:     filepath.Join(".cache", "whisper"),
			risk:        model.RiskRebuildable,
		},
		{
			id:          "nodegyp_cache",
			title:       i18n.T("node-gyp Header Cache", "node-gyp 原生模块编译头文件缓存"),
			description: i18n.T("Node.js C++ header files for compiling native addons", "编译 Node 原生扩展下载的 C++ 头文件缓存"),
			subpath:     filepath.Join("Library", "Caches", "node-gyp"),
			risk:        model.RiskRebuildable,
		},
	}

	// clang 模块缓存位于每用户缓存目录 ($TMPDIR 同级的 C/)，Xcode / swiftc / clang 编译时按需重建
	if tmp := filepath.Clean(os.TempDir()); strings.HasPrefix(tmp, "/var/folders/") || strings.HasPrefix(tmp, "/private/var/folders/") {
		targets = append(targets, cacheTarget{
			id:          "clang_module_cache",
			title:       i18n.T("Clang Module Cache", "Clang 预编译模块缓存"),
			description: i18n.T("Precompiled module cache shared by Xcode, swiftc and clang; rebuilt on the next compile", "Xcode / swiftc / clang 共用的预编译模块缓存，下次编译时自动重建"),
			absPath:     filepath.Join(filepath.Dir(tmp), "C", "clang", "ModuleCache"),
			risk:        model.RiskRebuildable,
		})
	}

	group := &model.Group{
		ID:       "package_caches",
		Title:    i18n.T("📦 Package Manager & Developer Caches", "📦 开发语言与包管理器缓存"),
		Category: model.CategoryPackageCache,
		Items:    make([]*model.Item, 0),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, tgt := range targets {
		tgt := tgt
		if len(tgt.subpaths) > 0 {
			var paths []string
			for _, sp := range tgt.subpaths {
				paths = append(paths, filepath.Join(home, sp))
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				item := newCacheItem(tgt.id, tgt.title, tgt.description, paths, 10*1024*1024, tgt.risk)
				if item == nil {
					return
				}
				item.Category = model.CategoryPackageCache
				mu.Lock()
				group.Items = append(group.Items, item)
				group.TotalSizeBytes += item.SizeBytes
				group.TotalReclaimableBytes += item.SizeBytes
				mu.Unlock()
			}()
			continue
		}
		fullPath := tgt.absPath
		if fullPath == "" {
			if tgt.subpath == "" || tgt.subpath == "." {
				continue
			}
			fullPath = filepath.Join(home, tgt.subpath)
		}
		if fullPath == home {
			continue
		}
		info, err := os.Stat(fullPath)
		if err != nil || !info.IsDir() {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}

			size := FastDirSize(fullPath)
			if size < 10*1024*1024 { // 小于 10MB 的微小缓存暂时忽略，保持输出干净
				return
			}

			item := &model.Item{
				ID:            tgt.id,
				Title:         tgt.title,
				Description:   tgt.description,
				Path:          fullPath,
				SizeBytes:     size,
				SizeFormatted: model.FormatBytes(size),
				Risk:          tgt.risk,
				Category:      model.CategoryPackageCache,
				IsProtected:   false,
				CleanType:     model.CleanTypeRemovePath,
				CleanPath:     fullPath,
			}

			if len(tgt.cleanCmd) > 0 {
				item.CleanType = model.CleanTypeCommand
				item.CleanCommand = tgt.cleanCmd
			}

			mu.Lock()
			group.Items = append(group.Items, item)
			group.TotalSizeBytes += size
			group.TotalReclaimableBytes += size
			mu.Unlock()
		}()
	}

	wg.Wait()

	return group, nil
}
