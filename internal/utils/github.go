package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/ui"
)

// GetLatestReleaseTag 查询 GitHub 仓库的最新 release tag
// repoFullName 格式示例："docker/buildx"
// proxy 为可选代理前缀，例如 "https://ghproxy.com/"
// 如果不需要代理，传空字符串 "" 即可。
func GetLatestReleaseTag(ui ui.UI, repoFullName, githubProxy, httpProxy string) (string, error) {
	if repoFullName == "" || !strings.Contains(repoFullName, "/") {
		return "", fmt.Errorf("invalid repoFullName: %s, expected format 'owner/repo'", repoFullName)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repoFullName)
	if httpProxy == "" {
		apiURL = ProxyURL(ui, githubProxy, apiURL)
	}

	// 设置 HTTP 客户端
	client := &http.Client{}
	if httpProxy != "" {
		proxyURL, err := url.Parse(httpProxy)
		if err != nil {
			ui.Error("Invalid HTTP proxy URL: %v", err)
			return "", err
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Go-http-client") // GitHub API 要求 UA

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get latest release: HTTP %d", resp.StatusCode)
	}

	var body struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}

	if body.TagName == "" {
		return "", fmt.Errorf("tag_name not found in release info")
	}

	return body.TagName, nil
}

// proxyURL 添加GitHub代理前缀
func ProxyURL(ui ui.UI, githubProxy, url string) string {
	if githubProxy != "" {
		if !strings.HasSuffix(githubProxy, "/") {
			githubProxy += "/"
		}
		ui.Info("使用代理访问: %s -> %s%s", url, githubProxy, url)
		return fmt.Sprintf("%s%s", githubProxy, url)
	}
	return url
}

// isGitRepo 检查目录是否是有效的git仓库
func IsGitRepo(ctx context.Context, ui ui.UI, path string) bool {
	ui.Info("校验Git仓库%s状态...", path)
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		ui.Info("路径 %s 不是Git仓库: 缺少.git目录", path)
		return false
	}

	// 验证git仓库状态
	err := RunCommand(ctx, ui, map[string]string{}, "git", "-C", path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		ui.Info("路径 %s 不是有效的Git仓库: %v", path, err)
		return false
	}

	ui.Info("路径 %s 是有效的Git仓库", path)
	return true
}

func CloneRepoWithProxy(
	ctx context.Context,
	ui ui.UI,
	repoURL string,
	path string,
	httpProxy string,
	githubProxy string,
	env map[string]string,
) error {
	if IsGitRepo(ctx, ui, path) {
		ui.Info("Git 仓库已存在: %s", path)
		return nil
	}

	// 目录存在但不是git仓库
	if PathExists(path) {
		ui.Warning("目录 %s 存在但不是git仓库，尝试备份", path)
		backupDir := path + ".bak"
		if err := os.Rename(path, backupDir); err != nil {
			return fmt.Errorf("备份目录失败: %w", err)
		}
		ui.Info("已备份到: %s", backupDir)
	}

	// 处理代理
	url := repoURL
	if httpProxy != "" {
		ui.Info("使用代理: %s", httpProxy)
		if env == nil {
			env = make(map[string]string)
		}
		env["HTTP_PROXY"] = httpProxy
		env["HTTPS_PROXY"] = httpProxy
	} else if githubProxy != "" {
		url = ProxyURL(ui, githubProxy, repoURL)
	}

	ui.Info("克隆仓库: %s -> %s", url, path)
	return RunCommand(ctx, ui, env, "git", "clone", "--depth=1", url, path)
}

// UpdateRepo 更新指定目录的 Git 仓库，可选 http/https 代理
func UpdateRepo(ctx context.Context, ui ui.UI, repoPath string, env map[string]string, httpProxy string) error {
	ui.Info("更新仓库: %s", repoPath)
	if !IsGitRepo(ctx, ui, repoPath) {
		return fmt.Errorf("目录 %s 不是有效的 git 仓库", repoPath)
	}

	// 设置代理
	if httpProxy != "" {
		ui.Info("使用代理: %s 更新仓库", httpProxy)
		if env == nil {
			env = make(map[string]string)
		}
		if env == nil {
			env = make(map[string]string)
		}
		env["HTTP_PROXY"] = httpProxy
		env["HTTPS_PROXY"] = httpProxy
	}

	ui.Info("更新仓库: %s", repoPath)
	return RunCommand(ctx, ui, env, "git", "-C", repoPath, "pull", "--ff-only")
}
